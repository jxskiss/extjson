package parser

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/tidwall/gjson"
)

//go:generate peg json.peg

const maxImportDepth = 10

func Parse(data []byte, importRoot string, enableEnv bool, funcMap map[string]interface{}) ([]byte, error) {
	return parse(data, importRoot, 0, enableEnv, funcMap)
}

func parse(data []byte, importRoot string, depth int, enableEnv bool, funcMap map[string]interface{}) ([]byte, error) {
	if depth > maxImportDepth {
		return nil, errors.New("max import depth exceeded")
	}

	doc := &JSON{
		Buffer: b2s(data),
	}
	if err := doc.Init(); err != nil {
		return nil, err
	}
	if err := doc.Parse(); err != nil {
		return nil, err
	}
	if !doc.hasExtendedFeature() {
		return data, nil
	}

	parser := &parser{
		doc:          doc,
		buf:          make([]byte, 0, len(data)),
		root:         importRoot,
		depth:        depth,
		enableEnv:    enableEnv,
		refTable:     make(map[string]int),
		inputFuncMap: funcMap,
		funcValMap:   make(map[string]reflect.Value),
	}
	parser.addFuncs(funcMap)

	return parser.rewrite()
}

type parser struct {
	doc *JSON
	buf []byte

	root  string
	depth int

	enableEnv bool

	refMark    string
	refCounter int
	refTable   map[string]int
	refDag     dag

	inputFuncMap map[string]interface{}
	funcValMap   map[string]reflect.Value
}

func (p *parser) text(n *node32) string {
	return p.doc.text(n.token32)
}

func (p *parser) rewrite() ([]byte, error) {
	root := p.doc.AST()
	if root.pegRule != ruleDocument {
		return nil, errors.New("invalid JSON document")
	}

	for n := root.up; n != nil; n = n.next {
		switch n.pegRule {
		case ruleSpacing:
			continue
		case ruleJSON:
			if err := p.parseJSON(n); err != nil {
				return nil, err
			}
		}
	}

	err := p.resolveReferences()
	if err != nil {
		return nil, err
	}

	return p.buf, nil
}

func (p *parser) resolveReferences() error {
	if len(p.refMark) == 0 {
		return nil
	}
	mark := p.refMark
	markLen := len(mark) + 1

	resolved := make(map[int]string)
	for path, seq := range p.refTable {
		r := gjson.GetBytes(p.buf, path)
		if !r.Exists() {
			return fmt.Errorf("cannot resolve reference %s", path)
		}

		resolved[seq] = r.Raw

		pos := 0
		for pos < len(r.Raw) {
			idx := strings.Index(r.Raw[pos:], mark)
			if idx < 0 {
				break
			}
			end := idx + markLen + 1
			for end < len(r.Raw) {
				if r.Raw[end] >= '0' && r.Raw[end] <= '9' {
					end++
					continue
				}
				break
			}
			refSeqStr := r.Raw[idx+markLen : end]
			refSeq, _ := strconv.ParseInt(refSeqStr, 10, 32)
			if int(refSeq) == seq {
				return fmt.Errorf("cannot reference to self %s", path)
			}
			p.refDag.addEdge(int(refSeq), seq)
			pos = end
		}
	}

	// resolving order
	order := p.refDag.sort()

	for _, seq := range order {
		final := resolved[seq]
		placeholder := `"` + p.referPlaceholder(seq) + `"`
		p.refDag.visitNeighbors(seq, func(to int) {
			resolved[to] = strings.Replace(resolved[to], placeholder, final, -1)
		})
	}
	oldnew := make([]string, 0, 2*len(resolved))
	for seq, text := range resolved {
		placeholder := `"` + p.referPlaceholder(seq) + `"`
		oldnew = append(oldnew, placeholder, text)
	}
	var repl = strings.NewReplacer(oldnew...)
	var buf bytes.Buffer
	repl.WriteString(&buf, b2s(p.buf))
	p.buf = buf.Bytes()
	return nil
}

func (p *parser) parseJSON(n *node32) (err error) {
	n = n.up
	switch n.pegRule {
	case ruleObject:
		if err = p.parseObject(n); err != nil {
			return
		}
	case ruleArray:
		if err = p.parseArray(n); err != nil {
			return
		}
	case ruleString:
		p.buf = append(p.buf, p.parseString(n, true)...)
	case ruleTrue:
		p.buf = append(p.buf, "true"...)
	case ruleFalse:
		p.buf = append(p.buf, "false"...)
	case ruleNull:
		p.buf = append(p.buf, "null"...)
	case ruleNumber:
		p.buf = append(p.buf, p.text(n)...)
	case ruleDirective:
		if err = p.parseDirective(n); err != nil {
			return
		}
	}
	return nil
}

func (p *parser) parseObject(n *node32) (err error) {
	var preRule pegRule
	for n := n.up; n != nil; n = n.next {
		switch n.pegRule {
		case ruleLWING:
			p.buf = append(p.buf, '{')
		case ruleRWING:
			if preRule == ruleCOMMA {
				p.buf = p.buf[:len(p.buf)-1]
			}
			p.buf = append(p.buf, '}')
		case ruleCOLON:
			p.buf = append(p.buf, ':')
		case ruleCOMMA:
			p.buf = append(p.buf, ',')
		case ruleObjectKey:
			p.buf = append(p.buf, p.parseObjectKey(n)...)
		case ruleJSON:
			err = p.parseJSON(n)
			if err != nil {
				return
			}
		}
		preRule = n.pegRule
	}
	return nil
}

func (p *parser) parseObjectKey(n *node32) string {
	n = n.up
	switch n.pegRule {
	case ruleSimpleIdentifier:
		return `"` + string(p.doc.buffer[n.begin:n.end]) + `"`
	case ruleString:
		return p.parseString(n, true)
	}
	return ""
}

func (p *parser) parseArray(n *node32) (err error) {
	var preRule pegRule
	for n := n.up; n != nil; n = n.next {
		switch n.pegRule {
		case ruleLBRK:
			p.buf = append(p.buf, '[')
		case ruleRBRK:
			if preRule == ruleCOMMA {
				p.buf = p.buf[:len(p.buf)-1]
			}
			p.buf = append(p.buf, ']')
		case ruleCOMMA:
			p.buf = append(p.buf, ',')
		case ruleJSON:
			err = p.parseJSON(n)
			if err != nil {
				return
			}
		}
		preRule = n.pegRule
	}
	return nil
}

var singleQuoteReplacer = strings.NewReplacer(`\'`, `'`, `"`, `\"`)

func (p *parser) parseString(n *node32, escapeDoubleQuote bool) string {
	n = n.up
	switch n.pegRule {
	case ruleSingleQuoteLiteral:
		text := string(p.doc.buffer[n.begin+1 : n.end-1])
		if escapeDoubleQuote {
			text = singleQuoteReplacer.Replace(text)
		}
		return `"` + text + `"`
	case ruleDoubleQuoteLiteral:
		return p.text(n)
	}
	return ""
}

func (p *parser) parseDirective(n *node32) (err error) {
	n = n.up
	switch n.pegRule {
	case ruleEnv:
		return p.parseEnv(n)
	case ruleInclude:
		return p.parseInclude(n)
	case ruleRefer:
		return p.parseRefer(n)
	case ruleFunc:
		return p.callFunction(n)
	}
	return nil
}

func (p *parser) parseEnv(n *node32) (err error) {
	if !p.enableEnv {
		return errors.New("env feature is not enabled")
	}
	n = n.up
	envName := p.parseString(n, true)
	envName = envName[1 : len(envName)-1]
	value := os.Getenv(envName)
	b, _ := json.Marshal(value)
	p.buf = append(p.buf, b...)
	return nil
}

func (p *parser) parseInclude(n *node32) (err error) {
	n = n.up
	importPath := p.parseString(n, true)
	importPath = filepath.Join(p.root, importPath[1:len(importPath)-1])
	included, err := ioutil.ReadFile(importPath)
	if err != nil {
		return
	}
	included, err = parse(included, p.root, p.depth+1, p.enableEnv, p.inputFuncMap)
	if err != nil {
		return
	}
	p.buf = append(p.buf, included...)
	return nil
}

func (p *parser) parseRefer(n *node32) (err error) {
	n = n.up
	jsonPath := p.parseString(n, true)
	jsonPath = jsonPath[1 : len(jsonPath)-1]
	seq, refId := p.getReferId(jsonPath)
	p.buf = append(p.buf, '"')
	p.buf = append(p.buf, refId...)
	p.buf = append(p.buf, '"')
	p.refDag.addVertex(seq)
	return nil
}

func (p *parser) getReferId(path string) (int, string) {
	if p.refMark == "" {
		for {
			mark := make([]byte, 16)
			_, err := rand.Read(mark)
			if err != nil {
				panic(fmt.Sprintf("rand.Read got error %v", err))
			}
			str := hex.EncodeToString(mark)
			if !strings.Contains(p.doc.Buffer, str) {
				p.refMark = str
				break
			}
		}
	}
	seq := p.refTable[path]
	if seq == 0 {
		p.refCounter++
		seq = p.refCounter
		p.refTable[path] = seq
	}
	placeholder := p.referPlaceholder(seq)
	return seq, placeholder
}

func (p *parser) referPlaceholder(n int) string {
	return fmt.Sprintf("%s:%d", p.refMark, n)
}

func (p *JSON) hasExtendedFeature() bool {
	var preRule pegRule
	for _, n := range p.Tokens() {
		switch n.pegRule {
		case ruleSingleQuoteLiteral,
			ruleDirective, ruleEnv, ruleInclude, ruleRefer, ruleFunc,
			ruleLongComment, ruleLineComment, rulePragma:
			return true
		case ruleRWING, ruleRBRK:
			if preRule == ruleCOMMA {
				return true
			}
		case ruleTrue:
			if p.text(n) != "true" {
				return true
			}
		case ruleFalse:
			if p.text(n) != "false" {
				return true
			}
		case ruleNull:
			if p.text(n) != "null" {
				return true
			}
		}
		preRule = n.pegRule
	}
	return false
}

func (p *JSON) text(n token32) string {
	return string(p.buffer[n.begin:n.end])
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
