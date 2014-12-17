package main

/*
   Copyright (C) 2014 Kouhei Maeda <mkouhei@palmtb.net>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

import (
	"fmt"
	"go/scanner"
	"go/token"
	"regexp"
	"strings"
	"sync/atomic"
)

type importSpec struct {
	imPath  string
	pkgName string
}

type signature struct {
	receiverID   string
	baseTypeName string
	params       string
	result       string
}

type funcDecl struct {
	name string
	sig  signature
	body []string
}

type methodSpecs struct {
	name string
	sig  signature
}

type typeDecl struct {
	typeID     string
	typeName   string
	fieldDecls []fieldDecl
	methSpecs  []methodSpecs
}

type fieldDecl struct {
	idList    string
	fieldType string
}

type parserSrc struct {
	brackets int32
	braces   int32
	paren    int32

	imPkgs    []importSpec
	funcDecls []funcDecl
	typeDecls []typeDecl
	body      []string
	main      []string

	imFlag      bool
	funcName    string
	typeFlag    string
	tFlag       bool
	mainFlag    bool
	preToken    token.Token
	preLit      string
	tmpFuncDecl funcDecl

	// 0: nofunc
	// 1: receiverID
	// 2: baseTypeName
	// 3: name
	// 4: params
	// 5: result
	// 6: body
	// 7: close
	// 8: close main
	posFuncSig int
}

func (p *parserSrc) putPackages(imPath, pkgName string, iq chan<- importSpec) {
	// put package to queue of `go get'
	if !searchPackage(importSpec{imPath, pkgName}, p.imPkgs) {
		is := importSpec{imPath, pkgName}
		p.imPkgs = append(p.imPkgs, is)
		iq <- is
	}
}

func searchPackage(pkg importSpec, pkgs []importSpec) bool {
	// search item from []string
	for _, l := range pkgs {
		if pkg.imPath == l.imPath && pkg.pkgName == l.pkgName {
			return true
		}
	}
	return false
}

func removeImportPackage(slice *[]importSpec, pkg importSpec) {
	s := *slice
	for i, item := range s {
		if item.imPath == pkg.imPath && item.pkgName == pkg.pkgName {
			s = append(s[:i], s[i+1:]...)
		}
	}
	*slice = s
}

func compareImportSpecs(A, B []importSpec) []importSpec {
	m := make(map[importSpec]int)
	for _, b := range B {
		m[b]++
	}
	var ret []importSpec
	for _, a := range A {
		if m[a] > 0 {
			m[a]--
			continue
		}
		ret = append(ret, a)
	}
	return ret
}

func (p *parserSrc) parserType(line string) bool {
	var pat string
	if p.typeFlag == "paren" || p.typeFlag == "struct" {
		pat = `\A[[:blank:]]*(((\w+)[[:blank:]]+(\S+)))()[[:blank:]]*\z`
	} else if p.typeFlag == "interface" {
		methName := `[[:blank:]]*(\((\w+[[:blank:]]+\*?\w+)\)[[:blank:]]*)?(\w+)[[:blank:]]*`
		params := `([\w_\*\[\],[:blank:]]+|[:blank:]*)`
		result := `\(([\w_\*\[\],[:blank:]]+)\)|([\w\*\[\][:blank:]]+)`
		pat = fmt.Sprintf(`\A%s\(%s\)[[:blank:]]*(%s)?[[:blank:]]*\z`, methName, params, result)
	} else {
		pat = `\A[[:blank:]]*type[[:blank:]]+((\()|(\w+)[[:blank:]]+((struct|interface)[[:blank:]]*\{|\S+)[[:blank:]]*)\z`
	}

	re := regexp.MustCompile(pat)
	num := re.NumSubexp()
	groups := re.FindAllStringSubmatch(line, num)
	// group[2]: "("
	// group[3]: identifier
	// group[4]: type (not struct, interface)
	// gropu[5]: "{"
	if len(groups) == 0 {
		return false
	}
	for _, group := range groups {
		if group[2] == "(" {
			p.typeFlag = "paren"
		} else if p.typeFlag == "" && (group[5] == "struct" || group[5] == "interface") {
			p.typeFlag = group[5]
			p.typeDecls = append(p.typeDecls, typeDecl{group[3], group[5], []fieldDecl{}, []methodSpecs{}})
		} else if p.typeFlag == "struct" {
			i := len(p.typeDecls) - 1
			p.typeDecls[i].fieldDecls = append(p.typeDecls[i].fieldDecls, fieldDecl{group[3], group[4]})
		} else if p.typeFlag == "interface" {
			// group[2]: MethodType (enable to define?)
			// group[3]: MethodName
			// group[4]: parameters
			// group[5]: returns

			methType := ""
			if group[2] != "" {
				methType = fmt.Sprintf("(%s)", group[2])
			}
			var result string
			if group[6] != "" {
				// multiple results
				result = fmt.Sprintf("(%s)", group[6])
			} else if group[5] != "" {
				// single result
				result = group[5]
			} else {
				result = ""
			}

			i := len(p.typeDecls) - 1
			p.typeDecls[i].methSpecs = append(p.typeDecls[i].methSpecs, methodSpecs{group[3], signature{"", methType, group[4], result}})
		} else {
			p.typeDecls = append(p.typeDecls, typeDecl{group[3], group[4], []fieldDecl{}, []methodSpecs{}})
		}
	}
	return true
}

func (p *parserSrc) parserTypeSpec(line string) bool {
	if p.typeFlag == "paren" {
		if strings.Contains(line, ")") && p.paren == 0 {
			p.typeFlag = ""
			return true
		}
	} else if p.typeFlag == "struct" || p.typeFlag == "interface" {
		if strings.Contains(line, "}") && p.braces == 0 {
			p.typeFlag = ""
			return true
		}
	}
	return false
}

func (p *parserSrc) parseLine(bline []byte, iq chan<- importSpec) bool {
	line := string(bline)
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(bline))
	s.Init(file, bline, nil, scanner.ScanComments)

	switch {
	case p.parserTypeSpec(line):
		// type spec of struct parser
		return false
	case p.parserType(line):
		// type parser
		return false
	}

	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		p.countBBP(tok)

		// ignore packageClause
		p.ignorePkg(tok)

		// parse import declare
		p.parseImPkg(tok, lit, iq)
		// parse func declare
		p.parseFunc(tok, lit)

		p.preToken = tok

	}

	if p.posFuncSig == 8 {
		p.posFuncSig = 0
		return true
	}

	return false
}

func convertImport(pkgs []importSpec) []string {
	// convert packages list to "import" statement
	l := []string{"import ("}
	for _, pkg := range pkgs {
		l = append(l, fmt.Sprintf(`%s "%s"`, pkg.pkgName, pkg.imPath))
	}
	return append(l, ")")
}

func (p *parserSrc) convertFuncDecls() []string {
	var lines []string
	for _, fun := range p.funcDecls {
		rcv := ""
		if fun.sig.receiverID != "" && fun.sig.baseTypeName != "" {
			rcv = fmt.Sprintf("(%s %s)", fun.sig.receiverID, fun.sig.baseTypeName)
		}

		if fun.sig.result == "" {
			lines = append(lines, fmt.Sprintf("func %s%s(%s) {", rcv, fun.name, fun.sig.params))
		} else {
			lines = append(lines, fmt.Sprintf("func %s%s(%s) (%s) {", rcv, fun.name, fun.sig.params, fun.sig.result))
		}

		for _, l := range fun.body {
			lines = append(lines, l)
		}
		lines = append(lines, "}")
	}
	return lines
}

func (p *parserSrc) convertTypeDecls() []string {
	l := []string{"type ("}
	for _, t := range p.typeDecls {
		sig := fmt.Sprintf("%s %s {", t.typeID, t.typeName)
		switch {
		case len(t.methSpecs) > 0:
			l = append(l, sig)
			for _, m := range t.methSpecs {
				sig = fmt.Sprintf("%s%s(%s)", m.sig.baseTypeName, m.name, m.sig.params)
				if m.sig.result != "" {
					sig = fmt.Sprintf("%s %s", sig, m.sig.result)
				}
				l = append(l, sig)
			}
			l = append(l, "}")
		case len(t.fieldDecls) > 0:
			l = append(l, sig)
			for _, f := range t.fieldDecls {
				if f.fieldType == "" {
					l = append(l, f.idList)
				} else {
					l = append(l, fmt.Sprintf("%s %s", f.idList, f.fieldType))
				}
			}
			l = append(l, "}")
		default:
			// rewrite sig
			sig = fmt.Sprintf("%s %s", t.typeID, t.typeName)
			l = append(l, sig)
		}
	}
	l = append(l, ")")
	return l
}

func (p *parserSrc) mergeLines() []string {
	// merge "package", "import", "func", "func main".
	l := []string{"package main"}
	l = append(l, convertImport(p.imPkgs)...)
	l = append(l, p.convertTypeDecls()...)
	l = append(l, p.convertFuncDecls()...)
	l = append(l, p.body...)
	l = append(l, "func main() {")
	l = append(l, p.main...)
	return append(l, "}")
}

func (p *parserSrc) countBBP(tok token.Token) {
	switch {
	case tok == token.LBRACE:
		// {
		atomic.AddInt32(&p.braces, 1)
	case tok == token.RBRACE:
		// }
		atomic.AddInt32(&p.braces, -1)
	case tok == token.LBRACK:
		// [
		atomic.AddInt32(&p.brackets, 1)
	case tok == token.RBRACK:
		// ]
		atomic.AddInt32(&p.brackets, -1)
	case tok == token.LPAREN:
		// (
		atomic.AddInt32(&p.paren, 1)
	case tok == token.RPAREN:
		// )
		atomic.AddInt32(&p.paren, -1)
	}
}

func (p *parserSrc) validateBBP() bool {
	switch {
	case p.braces < 0, p.brackets < 0, p.paren < 0:
		return false
	default:
		return true
	}
}

func (p *parserSrc) ignorePkg(tok token.Token) bool {
	switch {
	case tok == token.PACKAGE, tok == token.IDENT && p.preToken == token.PACKAGE:
		p.preToken = tok
	default:
		return false
	}
	return true
}

func rmQuot(lit string) string {
	pat := `"(.|\S+|[\S/]+)"`
	re := regexp.MustCompile(pat)
	grp := re.FindStringSubmatch(lit)
	if len(grp) == 0 {
		return lit
	}
	return grp[1]
}

func (p *parserSrc) parseImPkg(tok token.Token, lit string, iq chan<- importSpec) bool {
	switch {
	case tok == token.IMPORT:
		p.imFlag = true
		p.preToken = tok
	case p.imFlag:
		switch {
		case tok == token.IDENT:
			p.preLit = lit
		case tok == token.STRING:
			p.putPackages(rmQuot(lit), p.preLit, iq)
			p.preLit = ""
		case tok == token.SEMICOLON && p.paren == 0:
			p.imFlag = false
			p.preToken = tok
		}
	default:
		return false
	}
	return true
}

func (p *parserSrc) parseFunc(tok token.Token, lit string) bool {
	str := tokenToStr(tok, lit)

	switch {
	case p.posFuncSig == 0:
		if tok == token.FUNC {
			p.posFuncSig = 1
			p.preLit = ""
		}
	case p.posFuncSig == 1 && p.paren > 0:
		// receiverID
		// func (ri rt) fname(pi pt) (res)
		//       ~~
		if tok == token.IDENT {
			p.tmpFuncDecl.sig.receiverID = str
			p.posFuncSig = 2
		}
	case p.posFuncSig == 2:
		// baseTypeName
		p.parseFuncBaseTypeName(tok, str)

	case p.posFuncSig == 3, p.posFuncSig == 1 && p.paren == 0:
		// funcName
		p.parseFuncName(tok, str)

	case p.posFuncSig == 4:
		// params
		p.parseFuncParams(tok, lit)

	case p.posFuncSig == 5:
		// result
		p.parseFuncResult(tok, lit)

	case p.posFuncSig == 6:
		// body
		if p.mainFlag {
			p.parseFuncBody(&p.main, tok, str)
		} else {
			p.parseFuncBody(&p.tmpFuncDecl.body, tok, str)
		}

	case p.posFuncSig == 7:
		// closing
		p.funcClosing(tok)

	default:
		p.preLit = str
	}
	p.preToken = tok
	return true
}

func (p *parserSrc) parseFuncBaseTypeName(tok token.Token, lit string) {
	if p.paren > 0 && p.tmpFuncDecl.sig.receiverID != "" {
		switch {
		case tok == token.MUL, tok == token.LBRACK:
			// func (ri *rt) fname(pi pt) (res)
			//          ~
			// func (ri []rt) fname(pi pt) (res)
			//          ~
			p.tmpFuncDecl.sig.baseTypeName = lit
		case tok == token.RBRACK:
			// func (ri []rt) fname(pi pt) (res)
			//           ~
			if p.tmpFuncDecl.sig.baseTypeName == "[" {
				p.tmpFuncDecl.sig.baseTypeName += lit
			}
		case tok == token.IDENT:
			// func (ri rt) fname(pi pt) (res)
			//          ~~
			switch {
			case p.preToken == token.MUL && p.tmpFuncDecl.sig.baseTypeName == "*":
				// func (ri *rt) fname(pi pt) (res)
				//           ~~
				p.tmpFuncDecl.sig.baseTypeName += lit
			case p.preToken == token.RBRACK && p.tmpFuncDecl.sig.baseTypeName == "[]":
				// func (ri []rt) fname(pi pt) (res)
				//            ~~
				p.tmpFuncDecl.sig.baseTypeName += lit
			default:
				// func (ri rt) fname(pi pt) (res)
				//          ~~
				p.tmpFuncDecl.sig.baseTypeName = lit
			}
			p.posFuncSig = 3
		}
	}

}

func (p *parserSrc) parseFuncName(tok token.Token, lit string) {
	if p.paren == 0 && tok == token.IDENT && p.tmpFuncDecl.name == "" {
		// func (ri rt) fname(pi pt) (res)
		//              ~~~~~
		if lit == "main" {
			p.mainFlag = true
		} else {
			p.tmpFuncDecl.name = lit
			p.funcName = lit
		}
		p.posFuncSig = 4
	}
}

func (p *parserSrc) parseFuncParams(tok token.Token, lit string) {
	switch {
	case tok == token.IDENT:
		if p.tmpFuncDecl.sig.params == "" {
			// func (ri rt) fname(pi pt) (res)
			//                    ~~
			p.tmpFuncDecl.sig.params = lit
		} else {
			if p.preToken == token.COMMA {
				// func (ri rt) fname(pi pt, pi pt) (res)
				//                           ~~
				p.tmpFuncDecl.sig.params = fmt.Sprintf("%s, %s", p.tmpFuncDecl.sig.params, lit)
			} else if p.preToken == token.MUL || p.preToken == token.RBRACK {
				// func (ri rt) fname(pi *pt) (res)
				//                        ~
				// func (ri rt) fname(pi []pt) (res)
				//                         ~
				p.tmpFuncDecl.sig.params = fmt.Sprintf("%s%s", p.tmpFuncDecl.sig.params, lit)
			} else {
				// func (ri rt) fname(pi pt) (res)
				//                       ~~
				p.tmpFuncDecl.sig.params = fmt.Sprintf("%s %s", p.tmpFuncDecl.sig.params, lit)
			}
		}
	case tok == token.MUL, tok == token.LBRACK:
		if p.tmpFuncDecl.sig.params != "" {
			// func (ri rt) fname(pi *pt) (res)
			//                       ~
			// func (ri rt) fname(pi []pt) (res)
			//                       ~
			p.tmpFuncDecl.sig.params = fmt.Sprintf("%s %s", p.tmpFuncDecl.sig.params, lit)
		}
	case tok == token.RBRACK && p.preToken == token.LBRACK:
		// func (ri rt) fname(pi []pt) (res)
		//                        ~
		if p.tmpFuncDecl.sig.params != "" {
			p.tmpFuncDecl.sig.params = fmt.Sprintf("%s%s", p.tmpFuncDecl.sig.params, lit)
		}
	case tok == token.RPAREN && p.paren == 0:
		p.posFuncSig = 5
	case p.mainFlag && tok == token.RPAREN:
		p.posFuncSig = 6
	}

}

func (p *parserSrc) parseFuncResult(tok token.Token, lit string) {
	switch {
	case tok == token.IDENT:
		switch {
		case p.preToken == token.RPAREN:
			// func (ri rt) fname(pi pt) res
			//                           ~~~
			p.tmpFuncDecl.sig.result = lit
			p.posFuncSig = 6
		case p.preToken == token.LPAREN:
			// func (ri rt) fname(pi pt) (res, res)
			//                            ~~~
			p.tmpFuncDecl.sig.result = lit
		case p.tmpFuncDecl.sig.result != "":
			p.tmpFuncDecl.sig.result = fmt.Sprintf("%s, %s", p.tmpFuncDecl.sig.result, lit)
		}
	case p.paren == 0 && (tok == token.RPAREN || tok == token.LBRACE):
		p.posFuncSig = 6
	}

}

func (p *parserSrc) parseFuncBody(body *[]string, tok token.Token, lit string) {
	b := *body
	switch {
	case tok == token.SEMICOLON:
		b = append(b, p.preLit)
		p.preLit = ""
	case tok == token.LBRACE && p.braces == 1:
	case tok == token.RBRACE && p.braces == 0:
		p.posFuncSig = 7
	case p.preLit == "":
		p.preLit = lit
	case hasSpaceToken(p.preToken) && hasSpaceToken(tok):
		p.preLit = fmt.Sprintf("%s %s", p.preLit, lit)
	default:
		p.preLit = fmt.Sprintf("%s%s", p.preLit, lit)
	}
	*body = b
}

func (p *parserSrc) funcClosing(tok token.Token) {
	if tok != token.IDENT && p.paren == 0 {
		if p.mainFlag {
			p.posFuncSig = 8
		} else {
			p.funcDecls = append(p.funcDecls, p.tmpFuncDecl)
			p.posFuncSig = 0
		}
		p.tmpFuncDecl = funcDecl{}
		p.mainFlag = false
	}

}

func tokenToStr(tok token.Token, lit string) string {
	var str string
	switch {
	case tok == token.ADD:
		str = "+"
	case tok == token.SUB:
		str = "-"
	case tok == token.MUL:
		str = "*"
	case tok == token.QUO:
		str = "/"
	case tok == token.REM:
		str = "%"
	case tok == token.AND:
		str = "&"
	case tok == token.OR:
		str = "|"
	case tok == token.XOR:
		str = "^"
	case tok == token.SHL:
		str = "<<"
	case tok == token.SHR:
		str = ">>"
	case tok == token.AND_NOT:
		str = "&^"
	case tok == token.ADD_ASSIGN:
		str = "+="
	case tok == token.SUB_ASSIGN:
		str = "-="
	case tok == token.MUL_ASSIGN:
		str = "*="
	case tok == token.QUO_ASSIGN:
		str = "/="
	case tok == token.REM_ASSIGN:
		str = "%="
	case tok == token.AND_ASSIGN:
		str = "&="
	case tok == token.OR_ASSIGN:
		str = "|="
	case tok == token.XOR_ASSIGN:
		str = "^="
	case tok == token.SHL_ASSIGN:
		str = "<<="
	case tok == token.SHR_ASSIGN:
		str = ">>="
	case tok == token.AND_NOT_ASSIGN:
		str = "&^="
	case tok == token.LAND:
		str = "&&"
	case tok == token.LOR:
		str = "||"
	case tok == token.ARROW:
		str = "<-"
	case tok == token.INC:
		str = "++"
	case tok == token.DEC:
		str = "--"
	case tok == token.EQL:
		str = "=="
	case tok == token.LSS:
		str = "<"
	case tok == token.GTR:
		str = ">"
	case tok == token.ASSIGN:
		str = "="
	case tok == token.NOT:
		str = "!"
	case tok == token.NEQ:
		str = "!="
	case tok == token.LEQ:
		str = "<="
	case tok == token.GEQ:
		str = ">="
	case tok == token.DEFINE:
		str = ":="
	case tok == token.ELLIPSIS:
		str = "..."
	case tok == token.LPAREN:
		str = "("
	case tok == token.LBRACK:
		str = "["
	case tok == token.LBRACE:
		str = "{"
	case tok == token.COMMA:
		str = ","
	case tok == token.PERIOD:
		str = "."
	case tok == token.RPAREN:
		str = ")"
	case tok == token.RBRACK:
		str = "]"
	case tok == token.RBRACE:
		str = "}"
	case tok == token.SEMICOLON:
		str = ";"
	case tok == token.COLON:
		str = ":"
	case tok == token.BREAK:
		str = "break"
	case tok == token.CASE:
		str = "case"
	case tok == token.CHAN:
		str = "chan"
	case tok == token.CONST:
		str = "const"
	case tok == token.CONTINUE:
		str = "continue"
	case tok == token.DEFAULT:
		str = "default"
	case tok == token.DEFER:
		str = "defer"
	case tok == token.ELSE:
		str = "else"
	case tok == token.FALLTHROUGH:
		str = "fallthrough"
	case tok == token.FOR:
		str = "for"
	case tok == token.FUNC:
		str = "func"
	case tok == token.GO:
		str = "go"
	case tok == token.GOTO:
		str = "goto"
	case tok == token.IF:
		str = "if"
	case tok == token.IMPORT:
		str = "import"
	case tok == token.INTERFACE:
		str = "interface"
	case tok == token.MAP:
		str = "map"
	case tok == token.PACKAGE:
		str = "package"
	case tok == token.RANGE:
		str = "range"
	case tok == token.RETURN:
		str = "return"
	case tok == token.SELECT:
		str = "select"
	case tok == token.STRUCT:
		str = "struct"
	case tok == token.SWITCH:
		str = "switch"
	case tok == token.TYPE:
		str = "type"
	case tok == token.VAR:
		str = "var"
	default:
		str = lit
	}
	return str
}

func hasSpaceToken(tok token.Token) bool {
	switch {
	case tok == token.LBRACK:
	case tok == token.RBRACK:
	case tok == token.BREAK:
	case tok == token.CASE:
	case tok == token.CHAN:
	case tok == token.CONST:
	case tok == token.CONTINUE:
	case tok == token.DEFAULT:
	case tok == token.DEFER:
	case tok == token.ELSE:
	case tok == token.FALLTHROUGH:
	case tok == token.FOR:
	case tok == token.FUNC:
	case tok == token.GO:
	case tok == token.GOTO:
	case tok == token.IF:
	case tok == token.IMPORT:
	case tok == token.INTERFACE:
	case tok == token.MAP:
	case tok == token.PACKAGE:
	case tok == token.RANGE:
	case tok == token.RETURN:
	case tok == token.SELECT:
	case tok == token.STRUCT:
	case tok == token.SWITCH:
	case tok == token.TYPE:
	case tok == token.VAR:
	case tok == token.IDENT:
	case tok == token.INT:
	case tok == token.FLOAT:
	case tok == token.IMAG:
	case tok == token.CHAR:
	case tok == token.STRING:
	default:
		return false
	}
	return true
}
