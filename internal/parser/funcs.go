package parser

import (
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	gotoken "go/token"
	"io"
	mrand "math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

var (
	rngMu sync.Mutex
	_rng  = mrand.New(mrand.NewSource(time.Now().UnixNano()))
)

func (p *parser) addFuncs(funcMap map[string]interface{}) {
	for name, fn := range builtinFuncs {
		p.funcValMap[name] = reflect.ValueOf(fn)
	}
	for name, fn := range funcMap {
		p.funcValMap[name] = reflect.ValueOf(fn)
	}
}

func (p *parser) callFunction(n *node32) (err error) {
	n = n.up
	str := p.parseString(n, false)
	str = str[1 : len(str)-1]
	str = strings.Replace(str, `'`, `"`, -1)

	var funcName string
	var fn reflect.Value
	var callArgs []reflect.Value

	if fn = p.funcValMap[str]; fn.IsValid() {
		funcName = str
	} else {
		var expr *expression
		expr, err = parseExpression(str)
		if err != nil {
			return err
		}
		fn = p.funcValMap[expr.Func]
		if !fn.IsValid() {
			return fmt.Errorf("function %s is unknown", expr.Func)
		}
		fnTyp := fn.Type()
		if len(expr.Args) != fnTyp.NumIn() {
			return fmt.Errorf("function %s arguments count not match", expr.Func)
		}
		funcName = expr.Func
		for i := 0; i < len(expr.Args); i++ {
			fnArgTyp := fnTyp.In(i)
			if !expr.Args[i].Type().ConvertibleTo(fnArgTyp) {
				return fmt.Errorf("function %s argument type not match: %v", expr.Func, expr.Args[i].Interface())
			}
			callArgs = append(callArgs, expr.Args[i].Convert(fnArgTyp))
		}
	}

	out := fn.Call(callArgs)
	if len(out) > 1 && !out[1].IsNil() {
		return fmt.Errorf("call function %q: %w", funcName, out[1].Interface().(error))
	}

	result := out[0]
	if result.Kind() == reflect.String {
		ret0 := result.String()
		p.buf = append(p.buf, '"')
		p.buf = append(p.buf, strings.Replace(ret0, `"`, `\"`, -1)...)
		p.buf = append(p.buf, '"')
	} else {
		p.buf = append(p.buf, fmt.Sprint(result.Interface())...)
	}
	return nil
}

var (
	errNotCallExpression        = errors.New("not a call expression")
	errArgumentTypeNotSupported = errors.New("argument type not supported")
)

type expression struct {
	Func string
	Args []reflect.Value
}

func parseExpression(str string) (*expression, error) {
	expr, err := goparser.ParseExpr(str)
	if err != nil {
		return nil, err
	}
	call, ok := expr.(*goast.CallExpr)
	if !ok {
		return nil, errNotCallExpression
	}
	fnName := call.Fun.(*goast.Ident).String()
	var args []reflect.Value
	for _, a := range call.Args {
		lit, ok := a.(*goast.BasicLit)
		if !ok {
			return nil, errArgumentTypeNotSupported
		}
		var aVal interface{}
		switch lit.Kind {
		case gotoken.INT:
			aVal, _ = strconv.ParseInt(lit.Value, 10, 64)
		case gotoken.FLOAT:
			aVal, _ = strconv.ParseFloat(lit.Value, 64)
		case gotoken.STRING:
			aVal = lit.Value[1 : len(lit.Value)-1]
		default:
			return nil, errArgumentTypeNotSupported
		}
		args = append(args, reflect.ValueOf(aVal))
	}
	return &expression{
		Func: fnName,
		Args: args,
	}, nil
}

// -------- builtins -------- //

var builtinFuncs = map[string]interface{}{
	"nowUnix":    builtinNowUnix,
	"nowMilli":   builtinNowMilli,
	"nowNano":    builtinNowNano,
	"nowRFC3339": builtinNowRFC3339,
	"nowFormat":  builtinNowFormat,
	"uuid":       builtinUuid,
	"rand":       builtinRand,
	"randN":      builtinRandN,
	"randStr":    builtinRandStr,
}

func builtinNowUnix() int64 {
	return time.Now().Unix()
}

func builtinNowMilli() int64 {
	return time.Now().UnixNano() / 1e6
}

func builtinNowNano() int64 {
	return time.Now().UnixNano()
}

func builtinNowRFC3339() string {
	return time.Now().Format(time.RFC3339)
}

func builtinNowFormat(layout string) string {
	return time.Now().Format(layout)
}

func builtinUuid() string {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(crand.Reader, uuid)
	if err != nil {
		panic(err)
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	buf := make([]byte, 36)
	hex.Encode(buf[:8], uuid[:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], uuid[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], uuid[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], uuid[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], uuid[10:])
	return *(*string)(unsafe.Pointer(&buf))
}

func builtinRand() (x int64) {
	rngMu.Lock()
	x = _rng.Int63()
	rngMu.Unlock()
	return
}

func builtinRandN(n int64) (x int64) {
	rngMu.Lock()
	x = _rng.Int63n(n)
	rngMu.Unlock()
	return
}

func builtinRandStr(n int) string {
	const table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	buf := make([]byte, n)

	rngMu.Lock()
	defer rngMu.Unlock()
	for i := range buf {
		buf[i] = table[_rng.Intn(len(table))]
	}
	return *(*string)(unsafe.Pointer(&buf))
}
