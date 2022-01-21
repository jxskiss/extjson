package parser

import (
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"io"
	mrand "math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

func init() {
	mrand.Seed(time.Now().UnixNano())
}

func (p *parser) callFunction(n *node32) (err error) {
	n = n.up
	str := p.parseString(n, false)
	str = str[1 : len(str)-1]
	str = strings.Replace(str, `'`, `"`, -1)

	var fn interface{}
	var callArgs []reflect.Value

	if fn = builtinFuncs[str]; fn != nil {
		// pass
	} else {
		var expr *expression
		expr, err = parseExpression(str)
		if err != nil {
			return err
		}
		fn = builtinFuncs[expr.Func]
		if fn == nil {
			return fmt.Errorf("function %s is unknown", expr.Func)
		}
		fnVal := reflect.ValueOf(fn)
		fnTyp := fnVal.Type()
		if len(expr.Args) != fnTyp.NumIn() {
			return fmt.Errorf("function %s arguments count not match", expr.Func)
		}
		for i := 0; i < len(expr.Args); i++ {
			fnArgTyp := fnTyp.In(i)
			if !expr.Args[i].Type().ConvertibleTo(fnArgTyp) {
				return fmt.Errorf("function %s argument type not match: %v", expr.Func, expr.Args[i].Interface())
			}
			callArgs = append(callArgs, expr.Args[i].Convert(fnArgTyp))
		}
	}

	out := reflect.ValueOf(fn).Call(callArgs)
	switch ret0 := out[0].Interface().(type) {
	case string:
		p.buf = append(p.buf, '"')
		p.buf = append(p.buf, strings.Replace(ret0, `"`, `\"`, -1)...)
		p.buf = append(p.buf, '"')
	default:
		p.buf = append(p.buf, fmt.Sprint(ret0)...)
	}
	return nil
}

var (
	errNotCallExpression        = errors.New("not a call expression")
	errArgumentTypeNotSupported = errors.New("argument type not supported")
)

type FuncMap map[string]interface{}

type expression struct {
	Func string
	Args []reflect.Value
}

func parseExpression(str string) (*expression, error) {
	expr, err := goparser.ParseExpr(str)
	if err != nil {
		return nil, err
	}
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil, errNotCallExpression
	}
	fnName := call.Fun.(*ast.Ident).String()
	var args []reflect.Value
	for _, a := range call.Args {
		lit, ok := a.(*ast.BasicLit)
		if !ok {
			return nil, errArgumentTypeNotSupported
		}
		var aVal interface{}
		switch lit.Kind {
		case token.INT:
			aVal, _ = strconv.ParseInt(lit.Value, 10, 64)
		case token.FLOAT:
			aVal, _ = strconv.ParseFloat(lit.Value, 64)
		case token.STRING:
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

var builtinFuncs = FuncMap{
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

func builtinRand() int64 {
	return mrand.Int63()
}

func builtinRandN(n int64) int64 {
	return mrand.Int63n(n)
}

func builtinRandStr(n int) string {
	const table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	buf := make([]byte, n)
	max := len(table)
	for i := range buf {
		buf[i] = table[mrand.Intn(max)]
	}
	return *(*string)(unsafe.Pointer(&buf))
}
