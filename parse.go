package main

/*
   Copyright (C) 2014,2015 Kouhei Maeda <mkouhei@palmtb.net>

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
	mainHist  []string

	imFlag      bool
	funcName    string
	typeFlag    string
	tFlag       bool
	mainFlag    bool
	preToken    token.Token
	preLit      string
	tmpFuncDecl funcDecl
	tmpTypeDecl typeDecl

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

	// 0: notype
	// 1: typeID
	// 2: typeName
	// 3: fieldIDList
	// 4: fieldType
	// 5: method
	// 6: closing
	posType int

	// 0: no method
	// 1: method name
	// 2: param typeID
	// 3: param typeName
	// 4: result
	posMeth int
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

func (p *parserSrc) parseType(tok token.Token, lit string) bool {
	switch {
	case p.posType == 0 && tok == token.TYPE:
		// type typeID typeName
		// ~~~~
		//p.tFlag = true
		p.posType = 1
	case p.posType == 1:
		p.parseTypeID(tok, lit)
	case p.posType == 2:
		p.parseTypeName(tok, lit)
	case p.posType == 3:
		p.parseStructTypeID(tok, lit)
	case p.posType == 4:
		p.parseStructTypeName(tok, lit)
	case p.posType == 5:
		p.parseInterface(tok, lit)
	case p.posType == 6 && tok == token.SEMICOLON:
		p.posType = 0
		p.preLit = ""
	default:
		return false
	}
	return true
}

func (p *parserSrc) parseTypeID(tok token.Token, lit string) bool {
	// type typeID typeName
	//      ~~~~~~
	switch {
	case tok == token.IDENT && p.tmpTypeDecl.typeID == "":
		p.tmpTypeDecl.typeID = lit
		p.posType = 2
	case tok == token.RPAREN && p.paren == 0:
		p.posType = 6
	default:
		return false
	}
	return true
}

func (p *parserSrc) parseTypeName(tok token.Token, lit string) bool {
	if p.tmpTypeDecl.typeID != "" {
		switch {
		case tok == token.LBRACK:
			// type typeID []typeName
			//             ~
			p.tmpTypeDecl.typeName = lit
		case tok == token.RBRACK:
			// type typeID []typeName
			//              ~
			if p.tmpTypeDecl.typeName == "[" {
				p.tmpTypeDecl.typeName += lit
			}
		case tok == token.STRUCT:
			// type typeID struct {
			//             ~~~~~~
			p.tmpTypeDecl.typeName = lit
			p.posType = 3
		case tok == token.INTERFACE:
			// type typeID interface {
			//             ~~~~~~~~~
			p.tmpTypeDecl.typeName = lit
			p.posType = 5
			p.posMeth = 1
		case tok == token.IDENT:
			switch {
			case p.preToken == token.RBRACK && p.tmpTypeDecl.typeName == "[]":
				p.tmpTypeDecl.typeName += lit
				p.typeDecls = append(p.typeDecls, p.tmpTypeDecl)
				p.tmpTypeDecl = typeDecl{}
				if p.paren == 0 {
					// type typeID []typeName
					//               ~~~~~~~~
					p.posType = 6
				} else {
					// type (
					//    typeID []typeName
					//             ~~~~~~~~
					p.posType = 1
				}
			default:
				p.tmpTypeDecl.typeName = lit
				p.typeDecls = append(p.typeDecls, p.tmpTypeDecl)
				p.tmpTypeDecl = typeDecl{}
				if p.paren == 0 {
					// type typeID typeName
					//             ~~~~~~~~
					p.posType = 6
				} else {
					// type (
					//    typeID typeName
					//           ~~~~~~~~
					p.posType = 1
				}
			}
		default:
			return false
		}
		return true
	}
	return false
}

func (p *parserSrc) parseStructTypeID(tok token.Token, lit string) bool {
	// fieldIDList
	switch {
	case tok == token.RBRACE && p.braces == 0:
		if p.preToken == token.SEMICOLON {
			p.typeDecls = append(p.typeDecls, p.tmpTypeDecl)
			p.tmpTypeDecl = typeDecl{}
			p.preLit = ""
		}
		if p.paren == 0 {
			// type typeID struct {
			//     typeID []typeName
			//     }
			//     ~
			p.posType = 6
		} else if p.paren == 1 {
			// type (
			// typeID struct {
			//     typeID []typeName
			//     }
			//     ~
			p.posType = 1
		}
	case tok == token.IDENT && p.braces > 0:
		if p.preToken == token.LBRACE || p.preToken == token.SEMICOLON {
			// type typeID struct {
			//     typeID typeName
			//     ~~~~~~
			p.tmpTypeDecl.fieldDecls = append(p.tmpTypeDecl.fieldDecls, fieldDecl{lit, ""})
			p.posType = 4
		}
	default:
		return false
	}
	return true
}

func (p *parserSrc) parseStructTypeName(tok token.Token, lit string) bool {
	i := len(p.tmpTypeDecl.fieldDecls)
	if i > 0 {
		switch {
		case tok == token.IDENT && p.braces > 0:
			if p.tmpTypeDecl.fieldDecls[i-1].fieldType == "[]" {
				// type typeID struct {
				//     typeID []typeName
				//              ~~~~~~~~
				p.tmpTypeDecl.fieldDecls[i-1].fieldType += lit
			} else {
				// type typeID struct {
				//     typeID typeName
				//            ~~~~~~~~
				p.tmpTypeDecl.fieldDecls[i-1].fieldType = lit
			}
		case tok == token.LBRACK:
			// type typeID struct {
			//     typeID []typeName
			//            ~
			if p.tmpTypeDecl.fieldDecls[i-1].fieldType == "" {
				p.tmpTypeDecl.fieldDecls[i-1].fieldType = lit
			}
		case tok == token.RBRACK:
			// type typeID struct {
			//     typeID []typeName
			//             ~
			if p.tmpTypeDecl.fieldDecls[i-1].fieldType == "[" {
				p.tmpTypeDecl.fieldDecls[i-1].fieldType += lit
			}
		case tok == token.SEMICOLON:
			p.posType = 3
		default:
			return false
		}
		return true
	}
	return false
}

func (p *parserSrc) parseInterface(tok token.Token, lit string) bool {
	i := len(p.tmpTypeDecl.methSpecs)
	switch {
	case p.posMeth == 1:
		// type typeID interface {
		//     mname(pi pt) res
		//     ~~~~~
		switch {
		case tok == token.IDENT:
			p.tmpTypeDecl.methSpecs = append(p.tmpTypeDecl.methSpecs, methodSpecs{lit, signature{}})
		case tok == token.LPAREN:
			p.posMeth = 2
		case tok == token.RBRACE && p.braces == 0:
			p.typeDecls = append(p.typeDecls, p.tmpTypeDecl)
			p.tmpTypeDecl = typeDecl{}
			p.posMeth = 0
			if p.paren == 0 {
				p.posType = 6
			} else {
				p.posType = 1
			}
		}
	case p.posMeth == 2:
		switch {
		case tok == token.IDENT:
			// type typeID interface {
			//     mname(pi pt) res
			//           ~~
			if p.tmpTypeDecl.methSpecs[i-1].sig.params == "" {
				p.tmpTypeDecl.methSpecs[i-1].sig.params = lit
				p.posMeth = 3
			}
		case tok == token.RPAREN:
			p.posMeth = 4
		}
	case p.posMeth == 3:
		if p.tmpTypeDecl.methSpecs[i-1].sig.params != "" {
			switch {
			case tok == token.RPAREN:
				p.posMeth = 4
			case tok == token.LBRACK:
				// type typeID interface {
				//     mname(pi []pt) res
				//              ~
				p.tmpTypeDecl.methSpecs[i-1].sig.params += " " + lit
			case tok == token.RBRACK, tok == token.PERIOD:
				// type typeID interface {
				//     mname(pi []pt) res
				//               ~
				// or
				// type typeID interface {
				//     mname(pi pn.pt) res
				//                ~
				p.tmpTypeDecl.methSpecs[i-1].sig.params += lit
			case tok == token.IDENT:
				if strings.HasSuffix(p.tmpTypeDecl.methSpecs[i-1].sig.params, "[]") {
					// type typeID interface {
					//     mname(pi []pt) res
					//                ~~
					p.tmpTypeDecl.methSpecs[i-1].sig.params += lit
				} else if strings.HasSuffix(p.tmpTypeDecl.methSpecs[i-1].sig.params, ".") {
					// type typeID interface {
					//     mname(pi pn.pt) res
					//                 ~~
					p.tmpTypeDecl.methSpecs[i-1].sig.params += lit
				} else {
					// type typeID interface {
					//     mname(pi pt) res
					//              ~~
					p.tmpTypeDecl.methSpecs[i-1].sig.params += " " + lit
				}
			}
		}
	case p.posMeth == 4:
		switch {
		case tok == token.IDENT:
			if p.preToken == token.COMMA {
				// type typeID interface {
				//     mname(pi pt) (res, res)
				//                        ~~~
				if p.tmpTypeDecl.methSpecs[i-1].sig.result != "" {
					p.tmpTypeDecl.methSpecs[i-1].sig.result += ", " + lit
				}
			} else {
				// type typeID interface {
				//     mname(pi pt) res
				//                  ~~~
				// or
				//     mname(pi pt) (res, res)
				//                   ~~~
				p.tmpTypeDecl.methSpecs[i-1].sig.result = lit
			}
		case tok == token.SEMICOLON:
			// type typeID interface {
			//     mname(pi pt)
			// or
			// type typeID interface {
			//     mname(pi pt) res
			// or
			// type typeID interface {
			//     mname(pi pt) (res, res)
			p.posMeth = 1
		}
	default:
		return false
	}
	return true
}

func (p *parserSrc) parseLine(bline []byte, iq chan<- importSpec) bool {
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(bline))
	s.Init(file, bline, nil, scanner.ScanComments)

	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		str := tokenToStr(tok, lit)
		p.countBBP(tok)

		switch {
		case p.ignorePkg(tok):
			// ignore packageClause

		case p.parseImPkg(tok, str, iq):
			// parse import declare

		case p.parseType(tok, str):
			// parse type

		case p.parseFunc(tok, str):
			// parse func declare

		default:
			// omit main
			p.parseOmit(&p.main, tok, str)
			if len(p.main) > 0 && p.validateMainBody() {
				p.mainHist = append(p.mainHist, p.main...)
				return true
			}
		}
		p.preToken = tok

		if p.posFuncSig == 8 {
			p.posFuncSig = 0
			return true
		}

	}

	return false
}

func (p *parserSrc) validateMainBody() bool {
	var str string
	var tp parserSrc
	for _, l := range p.main {
		str += l + "\n"
	}
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len([]byte(str)))
	s.Init(file, []byte(str), nil, scanner.ScanComments)
	for {
		_, tok, _ := s.Scan()
		if tok == token.EOF {
			break
		}
		tp.countBBP(tok)
		if tp.paren == 0 && tp.braces == 0 && tp.brackets == 0 && tok == token.SEMICOLON {
			return true
		}
	}
	return false
}

func convertImport(pkgs []importSpec) []string {
	// convert packages list to "import" statement
	if len(pkgs) == 0 {
		return []string{}
	}
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

		switch {
		case fun.sig.result == "":
			lines = append(lines, fmt.Sprintf("func %s%s(%s) {", rcv, fun.name, fun.sig.params))
		case strings.Contains(fun.sig.result, ","):
			lines = append(lines, fmt.Sprintf("func %s%s(%s) (%s) {", rcv, fun.name, fun.sig.params, fun.sig.result))
		default:
			lines = append(lines, fmt.Sprintf("func %s%s(%s) %s {", rcv, fun.name, fun.sig.params, fun.sig.result))
		}

		for _, l := range fun.body {
			lines = append(lines, l)
		}
		lines = append(lines, "}")
	}
	return lines
}

func (p *parserSrc) convertTypeDecls() []string {
	if len(p.typeDecls) == 0 {
		return []string{}
	}
	l := []string{"type ("}
	for _, t := range p.typeDecls {
		sig := fmt.Sprintf("%s %s {", t.typeID, t.typeName)
		switch {
		case len(t.methSpecs) > 0:
			l = append(l, sig)
			for _, m := range t.methSpecs {
				sig = fmt.Sprintf("%s%s(%s)", m.sig.baseTypeName, m.name, m.sig.params)
				if m.sig.result != "" {
					if strings.Index(m.sig.result, ",") == -1 {
						sig += " " + m.sig.result
					} else {
						sig += " (" + m.sig.result + ")"
					}
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
	if len(p.mainHist) > 0 {
		l = append(l, p.mainHist...)
	} else {
		l = append(l, p.main...)
	}
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
			var pl string
			if p.preLit == ";" {
				pl = ""
			} else {
				pl = p.preLit
			}
			p.putPackages(rmQuot(lit), pl, iq)
			p.preLit = ""
		case tok == token.SEMICOLON:
			if p.paren == 0 {
				p.imFlag = false
				p.preToken = tok
				p.preLit = ""
			}
		}
	default:
		return false
	}
	return true
}

func (p *parserSrc) parseFunc(tok token.Token, lit string) bool {
	switch {
	case p.posFuncSig == 0 && tok == token.FUNC:
		p.posFuncSig = 1
		p.preLit = ""

	case p.posFuncSig == 1 && p.paren > 0:
		// receiverID
		// func (ri rt) fname(pi pt) (res)
		//       ~~
		if tok == token.IDENT {
			p.tmpFuncDecl.sig.receiverID = lit
			p.posFuncSig = 2
		}
	case p.posFuncSig == 2:
		// baseTypeName
		p.parseFuncBaseTypeName(tok, lit)

	case p.posFuncSig == 3, p.posFuncSig == 1 && p.paren == 0:
		// funcName
		p.parseFuncName(tok, lit)

	case p.posFuncSig == 4:
		// params
		p.parseFuncParams(tok, lit)

	case p.posFuncSig == 5:
		// result
		p.parseFuncResult(tok, lit)

	case p.posFuncSig == 6:
		// body
		if p.mainFlag {
			p.parseFuncBody(&p.main, tok, lit)
		} else {
			p.parseFuncBody(&p.tmpFuncDecl.body, tok, lit)
		}

	case p.posFuncSig == 7:
		// closing
		p.funcClosing(tok)

	default:
		return false
	}
	p.preToken = tok
	return true
}

func (p *parserSrc) parseFuncBaseTypeName(tok token.Token, lit string) {
	if p.paren > 0 && p.tmpFuncDecl.sig.receiverID != "" {
		switch {
		case tok == token.MUL:
			// func (ri *rt) fname(pi pt) (res)
			//          ~
			p.tmpFuncDecl.sig.baseTypeName = lit
		case tok == token.IDENT:
			// func (ri rt) fname(pi pt) (res)
			//          ~~
			switch {
			case p.preToken == token.MUL && p.tmpFuncDecl.sig.baseTypeName == "*":
				// func (ri *rt) fname(pi pt) (res)
				//           ~~
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
			p.mainHist = nil
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
				p.tmpFuncDecl.sig.params += ", " + lit
			} else if p.preToken == token.MUL || p.preToken == token.RBRACK {
				// func (ri rt) fname(pi *pt) (res)
				//                        ~
				// func (ri rt) fname(pi []pt) (res)
				//                         ~
				p.tmpFuncDecl.sig.params += lit
			} else {
				// func (ri rt) fname(pi pt) (res)
				//                       ~~
				p.tmpFuncDecl.sig.params += " " + lit
			}
		}
	case tok == token.MUL, tok == token.LBRACK:
		if p.tmpFuncDecl.sig.params != "" {
			// func (ri rt) fname(pi *pt) (res)
			//                       ~
			// func (ri rt) fname(pi []pt) (res)
			//                       ~
			p.tmpFuncDecl.sig.params += " " + lit

		}
	case tok == token.RBRACK && p.preToken == token.LBRACK:
		// func (ri rt) fname(pi []pt) (res)
		//                        ~
		if p.tmpFuncDecl.sig.params != "" {
			p.tmpFuncDecl.sig.params += lit
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
		case p.preToken == token.PERIOD, p.preToken == token.MUL, strings.HasSuffix(p.tmpFuncDecl.sig.result, "[]"):
			p.tmpFuncDecl.sig.result += lit
		case p.tmpFuncDecl.sig.result != "":
			p.tmpFuncDecl.sig.result += ", " + lit
		}
	case tok == token.MUL, tok == token.LBRACK:
		switch {
		case p.preToken == token.RPAREN:
			// func (ri rt) fname(pi pt) *res
			//                           ~
			p.tmpFuncDecl.sig.result = lit
		case p.preToken == token.LPAREN:
			// func (ri rt) fname(pi pt) (*res, res)
			//                             ~~~
			p.tmpFuncDecl.sig.result = lit
		case p.tmpFuncDecl.sig.result != "":
			p.tmpFuncDecl.sig.result += ", " + lit
		}
	case tok == token.RBRACK && p.preToken == token.LBRACK, tok == token.PERIOD:
		if p.tmpFuncDecl.sig.result != "" {
			p.tmpFuncDecl.sig.result += lit
		}
	case p.paren == 0 && (tok == token.RPAREN || tok == token.LBRACE):
		p.posFuncSig = 6
	}

}

func (p *parserSrc) parseOmit(body *[]string, tok token.Token, lit string) {
	b := *body
	switch {
	case tok == token.SEMICOLON:
		if p.preLit != "" {
			b = append(b, p.preLit)
		}
		p.preLit = ""
	case p.preLit == "":
		p.preLit = lit
	case hasLineFeedAfter(tok), p.preToken == token.RPAREN && tok == token.LBRACE:
		p.preLit += lit
		b = append(b, p.preLit)
		p.preLit = ""
	case hasLineFeedBefore(tok):
		b = append(b, p.preLit)
		p.preLit = lit
	case hasSpaceBefore(p.preToken) && hasSpaceBefore(tok):
		p.preLit += " " + lit
	case hasSpaceAfter(tok):
		p.preLit += lit + " "
	default:
		p.preLit += lit
	}
	*body = b
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
	case hasLineFeedAfter(tok), p.preToken == token.RPAREN && tok == token.LBRACE:
		p.preLit += lit
		b = append(b, p.preLit)
		p.preLit = ""
	case hasLineFeedBefore(tok):
		b = append(b, p.preLit)
		p.preLit = lit
	case hasSpaceBefore(p.preToken) && hasSpaceBefore(tok):
		p.preLit += " " + lit
	case hasSpaceAfter(tok):
		p.preLit += lit + " "
	default:
		p.preLit += lit
	}
	*body = b
}

func (p *parserSrc) funcClosing(tok token.Token) {
	if p.mainFlag == true {
		p.posFuncSig = 8
	} else if tok != token.IDENT && p.paren == 0 {
		if i := p.searchFuncDecl(p.tmpFuncDecl.name); i != -1 {
			p.funcDecls[i].name = p.tmpFuncDecl.name
			p.funcDecls[i].sig = p.tmpFuncDecl.sig
			p.funcDecls[i].body = p.tmpFuncDecl.body
		} else {
			p.funcDecls = append(p.funcDecls, p.tmpFuncDecl)
		}
		p.tmpFuncDecl = funcDecl{}
		p.posFuncSig = 0
	}
	p.mainFlag = false
}

func (p *parserSrc) searchFuncDecl(name string) int {
	for i, fnc := range p.funcDecls {
		if fnc.name == name {
			return i
		}
	}
	return -1
}

func removePrintStmt(slice *[]string) {
	s := *slice
	var r []int
	for i, item := range s {
		if strings.Contains(item, "fmt.Print") {
			r = append(r, i)
		}
	}
	for i := len(r) - 1; i >= 0; i-- {
		s = append(s[:r[i]], s[r[i]+1:]...)
	}
	*slice = s
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

func hasSpaceBefore(tok token.Token) bool {
	switch {
	case tok == token.SHL:
	case tok == token.SHR:
	case tok == token.ADD_ASSIGN:
	case tok == token.SUB_ASSIGN:
	case tok == token.MUL_ASSIGN:
	case tok == token.QUO_ASSIGN:
	case tok == token.REM_ASSIGN:
	case tok == token.AND_ASSIGN:
	case tok == token.OR_ASSIGN:
	case tok == token.XOR_ASSIGN:
	case tok == token.SHL_ASSIGN:
	case tok == token.SHR_ASSIGN:
	case tok == token.AND_NOT_ASSIGN:
	case tok == token.LAND:
	case tok == token.LOR:
	case tok == token.LSS:
	case tok == token.GTR:
	case tok == token.ASSIGN:
	case tok == token.NOT:
	case tok == token.NEQ:
	case tok == token.LEQ:
	case tok == token.GEQ:
	case tok == token.EQL:
	case tok == token.DEFINE:
	case tok == token.LBRACK:
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

func hasSpaceAfter(tok token.Token) bool {
	switch {
	case tok == token.COMMA:
	default:
		return false
	}
	return true
}

func hasLineFeedAfter(tok token.Token) bool {
	switch {
	case tok == token.COLON:
	default:
		return false
	}
	return true
}

func hasLineFeedBefore(tok token.Token) bool {
	switch {
	case tok == token.RETURN:
	case tok == token.BREAK:
	default:
		return false
	}
	return true
}
