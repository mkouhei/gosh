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

type imptSpec struct {
	imPath  string
	pkgName string
}

type signature struct {
	recvID       string
	baseTypeName string
	params       string
	result       string
}

type funcDecl struct {
	name string
	sig  signature
	body []string
}

type methSpec struct {
	name string
	sig  signature
}

type typeDecl struct {
	typeID     string
	typeName   string
	fieldDecls fieldDecls
	methSpecs  methSpecs
}

type fieldDecl struct {
	idList    string
	fieldType string
}

type tokenLit struct {
	tok token.Token
	lit string
}

type cnt struct {
	brackets int32
	braces   int32
	paren    int32
}

type imPkgs []imptSpec
type funcDecls []funcDecl
type typeDecls []typeDecl
type fieldDecls []fieldDecl
type methSpecs []methSpec
type queue []tokenLit

type parserSrc struct {
	cnt cnt

	imPkgs    imPkgs
	funcDecls funcDecls
	typeDecls typeDecls
	body      []string
	main      []string
	mainHist  []string

	funcName    string
	tFlag       bool
	mainFlag    bool
	preToken    token.Token
	preLit      string
	tmpFuncDecl funcDecl
	tmpTypeDecl typeDecl

	queue queue

	// 0: nofunc
	// 1: recvID
	// 2: baseTypeName
	// 3: name
	// 4: params
	// 5: result
	// 6: body
	// 7: close
	// 8: close main
	posFuncSig int
}

func (q *queue) push(v tokenLit) {
	*q = append(*q, v)
}

func (q *queue) pop() tokenLit {
	ret := (*q)[len(*q)-1]
	*q = (*q)[0 : len(*q)-1]
	return ret
}

func (q *queue) dequeue() tokenLit {
	ret := (*q)[0]
	*q = (*q)[1:len(*q)]
	return ret
}

func (q *queue) checkLatestItem(tok token.Token) bool {
	if (*q)[len(*q)-1].tok == tok {
		return true
	}
	return false
}

func (q *queue) clear() {
	for len(*q) > 0 {
		q.pop()
	}
}

func (q *queue) checkQueueType(tok token.Token) bool {
	if len(*q) > 0 && (*q)[0].tok == tok {
		return true
	}
	return false
}

func (s *imPkgs) putPackages(imPath, pkgName string, imptQ chan<- imptSpec) {
	// put package to queue of `go get'
	if !searchPackage(imptSpec{imPath, pkgName}, *s) {
		*s = append(*s, imptSpec{imPath, pkgName})
		imptQ <- imptSpec{imPath, pkgName}
	}
}

func searchPackage(pkg imptSpec, pkgs []imptSpec) bool {
	// search item from []string
	for _, l := range pkgs {
		if pkg.imPath == l.imPath && pkg.pkgName == l.pkgName {
			return true
		}
	}
	return false
}

func (s *imPkgs) removeImport(msg string, pkg imptSpec) {
	// remove package from env.parser.imPkg
	if strings.Contains(msg,
		fmt.Sprintf(`package %s: unrecognized import path "%s"`,
			pkgName(pkg.pkgName, pkg.imPath),
			pkgName(pkg.pkgName, pkg.imPath))) {
		s.removeImportPackage(imptSpec{pkg.imPath, pkg.pkgName})
	}
}

func (s *imPkgs) removeImportPackage(pkg imptSpec) {
	for i, item := range *s {
		if item.imPath == pkg.imPath && item.pkgName == pkg.pkgName {
			*s = append((*s)[:i], (*s)[i+1:]...)
		}
	}
}

func compareImportSpecs(A, B []imptSpec) []imptSpec {
	m := make(map[imptSpec]int)
	for _, b := range B {
		m[b]++
	}
	var ret []imptSpec
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
	case len(p.queue) == 0 && tok == token.TYPE:
		p.queue.push(tokenLit{tok, lit})
	case p.queue.checkQueueType(token.TYPE):
		p.queue.push(tokenLit{tok, lit})

		if tok == token.SEMICOLON && p.cnt.paren == 0 && p.cnt.braces == 0 {
			c := cnt{}
			for len(p.queue) > 0 {
				t := p.queue.dequeue()
				p.storeTypeDecl(t.tok, t.lit, &c)
			}
		}
	default:
		return false
	}
	return true
}

func (p *parserSrc) storeTypeDecl(tok token.Token, lit string, c *cnt) {
	c.countAllBrackets(tok)

	switch {
	case p.tmpTypeDecl.setTypeID(tok, lit):
	case p.tmpTypeDecl.setTypeName(tok, lit):
	case p.tmpTypeDecl.setStruct(tok, lit, c):
	case p.tmpTypeDecl.setInterface(tok, lit, c):
	case tok == token.SEMICOLON:
		p.appendTypeDecl()
	}
}

func (t *typeDecl) setTypeID(tok token.Token, lit string) bool {
	if tok == token.IDENT && t.typeID == "" {
		t.typeID = lit
		return true
	}
	return false
}

func (t *typeDecl) setTypeName(tok token.Token, lit string) bool {
	switch {
	case t.typeName == "":
		if tok == token.IDENT || tok == token.LBRACK || tok == token.STRUCT || tok == token.INTERFACE {
			t.typeName = lit
		}
	case t.typeName == "[" && tok == token.RBRACK:
		t.typeName += lit
	case t.typeName == "[]" && tok == token.IDENT:
		t.typeName += lit
	default:
		return false
	}
	return true
}

func (t *typeDecl) setStruct(tok token.Token, lit string, c *cnt) bool {
	if t.typeName != "struct" {
		return false
	}
	n := len(t.fieldDecls)
	e := &fieldDecl{}
	if n == 0 {
		t.fieldDecls = append(t.fieldDecls, *e)
	} else {
		e = &t.fieldDecls[n-1]
	}
	switch {
	case e.idList == "" && tok == token.IDENT:
		e.idList = lit
	case strings.HasSuffix(e.idList, ", ") && tok == token.IDENT:
		e.idList += lit
	case e.idList != "" && tok == token.COMMA:
		e.idList += lit + " "
	case tok == token.LBRACK && e.fieldType == "":
		e.fieldType = lit
	case tok == token.RBRACK && e.fieldType == "[":
		e.fieldType += lit
	case e.fieldType == "[]" && tok == token.IDENT:
		e.fieldType += lit
	case tok == token.IDENT && e.fieldType == "" && e.idList != "":
		e.fieldType = lit
	case tok == token.SEMICOLON && c.braces > 0:
		t.fieldDecls = append(t.fieldDecls, fieldDecl{})
	case tok == token.RBRACE && c.braces == 0:
		t.fieldDecls = t.fieldDecls[0 : len(t.fieldDecls)-1]
	default:
		return false
	}
	return true
}

func (t *typeDecl) setInterface(tok token.Token, lit string, c *cnt) bool {
	if t.typeName != "interface" {
		return false
	}
	n := len(t.methSpecs)
	e := &methSpec{}
	if n == 0 {
		t.methSpecs = append(t.methSpecs, *e)
	} else {
		e = &t.methSpecs[n-1]
	}

	switch {
	case e.name == "" && tok == token.IDENT:
		e.name = lit
	case e.sig.params == "" && tok == token.LPAREN:
		e.sig.params = lit
	case isOpenedParen(e.sig.params):
		if e.sig.params == "(" && tok == token.IDENT {
			e.sig.params += lit + " "
		} else {
			e.sig.params += lit
		}
	case isClosedParan(e.sig.params) && tok == token.SEMICOLON:
		e.sig.params = e.sig.params[1 : len(e.sig.params)-1]
		if isClosedParan(e.sig.result) {
			e.sig.result = e.sig.result[1 : len(e.sig.result)-1]
		}
		t.methSpecs[n-1] = *e
		e = &methSpec{}
		t.methSpecs = append(t.methSpecs, *e)
	case isClosedParan(e.sig.params) && e.sig.result == "":
		e.sig.result += lit
	case isOpenedParen(e.sig.result):
		if tok == token.COMMA {
			e.sig.result += lit + " "
		} else {
			e.sig.result += lit
		}
	case tok == token.RBRACE && c.braces == 0:
		q := &t.methSpecs
		*q = (*q)[0 : len(*q)-1]
		t.methSpecs = *q
	default:
		return false
	}
	return true
}

func isOpenedParen(str string) bool {
	if strings.HasPrefix(str, "(") && !strings.HasSuffix(str, ")") {
		return true
	}
	return false
}

func isClosedParan(str string) bool {
	if strings.HasPrefix(str, "(") && strings.HasSuffix(str, ")") {
		return true
	}
	return false
}

func (p *parserSrc) parseLine(bline []byte, imptQ chan<- imptSpec) bool {
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

		p.cnt.countAllBrackets(tok)

		switch {
		case p.queue.ignorePkg(tok):
			// ignore packageClause

		case p.parseImPkg(tok, str, imptQ):
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
	var tp parserSrc
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len([]byte(concatLines(p.main, "\n"))))
	s.Init(file, []byte(concatLines(p.main, "\n")), nil, scanner.ScanComments)
	for {
		_, tok, _ := s.Scan()
		if tok == token.EOF {
			break
		}

		tp.cnt.countAllBrackets(tok)

		if tp.cnt.paren == 0 && tp.cnt.braces == 0 && tp.cnt.brackets == 0 && tok == token.SEMICOLON {
			return true
		}
	}
	return false
}

func (s *imPkgs) convertImport() []string {
	// convert packages list to "import" statement
	if len(*s) == 0 {
		return []string{}
	}
	l := []string{"import ("}
	for _, pkg := range *s {
		l = append(l, fmt.Sprintf(`%s "%s"`, pkg.pkgName, pkg.imPath))
	}
	return append(l, ")")
}

func (s *funcDecls) convertFuncDecls() []string {
	var lines []string
	for _, fun := range *s {
		rcv := ""
		if fun.sig.recvID != "" && fun.sig.baseTypeName != "" {
			rcv = fmt.Sprintf("(%s %s)", fun.sig.recvID, fun.sig.baseTypeName)
		}

		lines = append(lines, convertFuncSig(rcv, fun.name, fun.sig.params, fun.sig.result))
		lines = append(appendLines(fun.body, lines), "}")
	}
	return lines
}

func convertFuncSig(rcv, name, params, result string) string {
	var line string
	switch {
	case result == "":
		line = fmt.Sprintf("func %s%s(%s) {", rcv, name, params)
	case strings.Contains(result, ","):
		line = fmt.Sprintf("func %s%s(%s) (%s) {", rcv, name, params, result)
	default:
		line = fmt.Sprintf("func %s%s(%s) %s {", rcv, name, params, result)
	}
	return line
}

func (s *typeDecls) convertTypeDecls() []string {
	if len(*s) == 0 {
		return []string{}
	}
	l := []string{"type ("}
	for _, t := range *s {
		sig := fmt.Sprintf("%s %s {", t.typeID, t.typeName)
		switch {
		case len(t.methSpecs) > 0:
			t.methSpecs.convertMethSpecs(&l, sig)
		case len(t.fieldDecls) > 0:
			t.fieldDecls.convertFieldDecls(&l, sig)
		default:
			// rewrite sig
			sig = fmt.Sprintf("%s %s", t.typeID, t.typeName)
			l = append(l, sig)
		}
	}
	l = append(l, ")")
	return l
}

func (s *methSpecs) convertMethSpecs(lines *[]string, sig string) {
	l := *lines
	l = append(l, sig)
	for _, m := range *s {
		str := fmt.Sprintf("%s%s(%s)", m.sig.baseTypeName, m.name, m.sig.params)
		if m.sig.result != "" {
			if strings.Index(m.sig.result, ",") == -1 {
				str += " " + m.sig.result
			} else {
				str += " (" + m.sig.result + ")"
			}
		}
		l = append(l, str)
	}
	l = append(l, "}")
	*lines = l
}

func (s *fieldDecls) convertFieldDecls(lines *[]string, sig string) {
	l := *lines
	l = append(l, sig)
	for _, f := range *s {
		if f.fieldType == "" {
			l = append(l, f.idList)
		} else {
			l = append(l, fmt.Sprintf("%s %s", f.idList, f.fieldType))
		}
	}
	l = append(l, "}")
	*lines = l
}

func (p *parserSrc) mergeLines() []string {
	// merge "package", "import", "func", "func main".
	l := []string{"package main"}
	l = append(l, p.imPkgs.convertImport()...)
	l = append(l, p.typeDecls.convertTypeDecls()...)
	l = append(l, p.funcDecls.convertFuncDecls()...)
	l = append(l, p.body...)
	l = append(l, "func main() {")
	if len(p.mainHist) > 0 {
		l = append(l, p.mainHist...)
	} else {
		l = append(l, p.main...)
	}
	return append(l, "}")
}

func countBracket(c *int32, tok token.Token) {
	if tok == token.LBRACK {
		// [
		atomic.AddInt32(c, 1)
	} else if tok == token.RBRACK {
		// ]
		atomic.AddInt32(c, -1)
	}
}

func countBrace(c *int32, tok token.Token) {
	if tok == token.LBRACE {
		// [
		atomic.AddInt32(c, 1)
	} else if tok == token.RBRACE {
		// ]
		atomic.AddInt32(c, -1)
	}
}

func countParen(c *int32, tok token.Token) {
	if tok == token.LPAREN {
		// [
		atomic.AddInt32(c, 1)
	} else if tok == token.RPAREN {
		// ]
		atomic.AddInt32(c, -1)
	}
}

func (c *cnt) countAllBrackets(tok token.Token) {
	countParen(&c.paren, tok)
	countBrace(&c.braces, tok)
	countBracket(&c.brackets, tok)
}

func (q *queue) ignorePkg(tok token.Token) bool {
	switch {
	case tok == token.PACKAGE:
		q.push(tokenLit{tok, ""})
	case q.checkQueueType(token.PACKAGE):
		q.clear()
	default:
		return false
	}
	return true
}

func rmQuot(lit string) string {
	re := regexp.MustCompile(`"(.|\S+|[\S/]+)"`)
	if len(re.FindStringSubmatch(lit)) == 0 {
		return lit
	}
	return re.FindStringSubmatch(lit)[1]
}

func (p *parserSrc) parseImPkg(tok token.Token, lit string, imptQ chan<- imptSpec) bool {
	switch {
	case len(p.queue) == 0 && tok == token.IMPORT:
		p.queue.push(tokenLit{tok, lit})
	case p.queue.checkQueueType(token.IMPORT):
		switch {
		case tok == token.IDENT:
			p.queue.push(tokenLit{tok, lit})
		case tok == token.STRING:
			var s string
			if p.queue.checkLatestItem(token.IMPORT) {
				s = ""
			} else {
				s = p.queue.pop().lit
			}
			p.imPkgs.putPackages(rmQuot(lit), litSemicolon(s), imptQ)
		case tok == token.SEMICOLON:
			if p.cnt.paren == 0 {
				p.queue.clear()
			}
		}
	default:
		return false
	}
	return true
}

func litSemicolon(lit string) string {
	s := ""
	if lit == ";" {
		s = ""
	} else {
		s = lit
	}
	return s
}

func (p *parserSrc) parseFunc(tok token.Token, lit string) bool {
	switch {
	case p.posFuncSig == 0 && tok == token.FUNC:
		p.posFuncSig = 1
		p.preLit = ""

	case p.posFuncSig == 1 && p.cnt.paren > 0:
		// recvID
		// func (ri rt) fname(pi pt) (res)
		//       ~~
		if tok == token.IDENT {
			p.tmpFuncDecl.sig.recvID = lit
			p.posFuncSig = 2
		}
	case p.posFuncSig == 2:
		// baseTypeName
		p.parseFuncBaseTypeName(tok, lit)

	case p.posFuncSig == 3, p.posFuncSig == 1 && p.cnt.paren == 0:
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
	if p.cnt.paren > 0 && p.tmpFuncDecl.sig.recvID != "" {
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
	if p.cnt.paren == 0 && tok == token.IDENT && p.tmpFuncDecl.name == "" {
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
	case p.parseFuncParamsIdent(tok, lit):

	case tok == token.MUL, tok == token.LBRACK:
		if p.tmpFuncDecl.sig.params != "" {
			// func (ri rt) fname(pi *pt) (res)
			//                       ~
			// func (ri rt) fname(pi []pt) (res)
			//                       ~
			p.tmpFuncDecl.sig.params += " " + lit

		}
	case isSliceRBrack(tok, p.preToken):
		// func (ri rt) fname(pi []pt) (res)
		//                        ~
		if p.tmpFuncDecl.sig.params != "" {
			p.tmpFuncDecl.sig.params += lit
		}
	case p.isOutOfParen(tok):
		p.posFuncSig = 5
	case p.mainFlag && tok == token.RPAREN:
		p.posFuncSig = 6
	}
}

func (p *parserSrc) parseFuncParamsIdent(tok token.Token, lit string) bool {
	if tok != token.IDENT {
		return false
	}

	switch {
	case p.tmpFuncDecl.sig.params == "":
		// func (ri rt) fname(pi pt) (res)
		//                    ~~
		p.tmpFuncDecl.sig.params = lit
	case p.preToken == token.COMMA:
		// func (ri rt) fname(pi pt, pi pt) (res)
		//                           ~~
		p.tmpFuncDecl.sig.params += ", " + lit
	case p.preToken == token.MUL, p.preToken == token.RBRACK:
		// func (ri rt) fname(pi *pt) (res)
		//                        ~
		// func (ri rt) fname(pi []pt) (res)
		//                         ~
		p.tmpFuncDecl.sig.params += lit
	default:
		// func (ri rt) fname(pi pt) (res)
		//                       ~~
		p.tmpFuncDecl.sig.params += " " + lit
	}
	return true
}

func (p *parserSrc) parseFuncResult(tok token.Token, lit string) {
	switch {
	case tok == token.IDENT:
		p.parseFuncResutlType(lit)
	case tok == token.MUL, tok == token.LBRACK:
		p.parseFuncResultPointer(lit)
	case isSliceRBrack(tok, p.preToken), tok == token.PERIOD:
		if p.tmpFuncDecl.sig.result != "" {
			p.tmpFuncDecl.sig.result += lit
		}
	case p.isOutOfParen(tok), p.cnt.paren == 0 && tok == token.LBRACE:
		p.posFuncSig = 6
	}
}

func (p *parserSrc) parseFuncResutlType(lit string) {
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
}

func (p *parserSrc) parseFuncResultPointer(lit string) {
	switch {
	case p.preToken == token.RPAREN, p.preToken == token.LPAREN:
		// func (ri rt) fname(pi pt) *res | func (ri rt) fname(pi pt) (*res, res)
		//                           ~                                  ~~~
		p.tmpFuncDecl.sig.result = lit
	case p.tmpFuncDecl.sig.result != "":
		p.tmpFuncDecl.sig.result += ", " + lit
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
	case tok == token.LBRACE && p.cnt.braces == 1:
	case p.isOutOfBrace(tok):
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
	} else if tok != token.IDENT && p.cnt.paren == 0 {
		if i := p.funcDecls.searchFuncDecl(p.tmpFuncDecl.name); i != -1 {
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

func (s *funcDecls) searchFuncDecl(name string) int {
	for i, fnc := range *s {
		if fnc.name == name {
			return i
		}
	}
	return -1
}

func (s *typeDecls) searchTypeDecl(typeID string) int {
	for i, t := range *s {
		if t.typeID == typeID {
			return i
		}
	}
	return -1
}

func (p *parserSrc) appendTypeDecl() {
	if i := p.typeDecls.searchTypeDecl(p.tmpTypeDecl.typeID); i != -1 {
		p.typeDecls[i].typeID = p.tmpTypeDecl.typeID
		p.typeDecls[i].typeName = p.tmpTypeDecl.typeName
		p.typeDecls[i].fieldDecls = p.tmpTypeDecl.fieldDecls
		p.typeDecls[i].methSpecs = p.tmpTypeDecl.methSpecs
	} else {
		p.typeDecls = append(p.typeDecls, p.tmpTypeDecl)
	}
	p.tmpTypeDecl = typeDecl{}
}

func isSliceRBrack(tok, preToken token.Token) bool {
	if tok == token.RBRACK && preToken == token.LBRACK {
		return true
	}
	return false
}

func (p *parserSrc) isOutOfBrace(tok token.Token) bool {
	if tok == token.RBRACE && p.cnt.braces == 0 {
		return true
	}
	return false
}

func (p *parserSrc) isOutOfParen(tok token.Token) bool {
	if tok == token.RPAREN && p.cnt.paren == 0 {
		return true
	}
	return false
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
