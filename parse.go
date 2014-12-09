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
	"regexp"
	"strings"
	"sync/atomic"
)

type importSpec struct {
	imPath  string
	pkgName string
}

type signature struct {
	functype string
	params   string
	result   string
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
	pkgFlag   bool
	imPkgs    []importSpec
	imFlag    bool
	funcDecls []funcDecl
	funcFlag  string
	blackets  int32
	paren     int32
	typeDecls []typeDecl
	typeFlag  string
	body      []string
	mainFlag  bool
	main      []string
}

func (p *parserSrc) appendBody(line string) {
	for i, fun := range p.funcDecls {
		if p.funcFlag == fun.name {
			p.funcDecls[i].body = append(p.funcDecls[i].body, line)
		}
	}
}

func (p *parserSrc) bIncrement() {
	atomic.AddInt32(&p.blackets, 1)
}

func (p *parserSrc) bDecrement() {
	atomic.AddInt32(&p.blackets, -1)
}

func (p *parserSrc) pIncrement() {
	atomic.AddInt32(&p.paren, 1)
}

func (p *parserSrc) pDecrement() {
	atomic.AddInt32(&p.paren, -1)
}

func (p *parserSrc) countBlackets(line string) {
	switch {
	case strings.Contains(line, "{") && strings.Contains(line, "}"):
		for i := 0; i < strings.Count(line, "{"); i++ {
			p.bIncrement()
		}
		for i := 0; i < strings.Count(line, "}"); i++ {
			p.bDecrement()
		}
	case strings.Contains(line, "{"):
		for i := 0; i < strings.Count(line, "{"); i++ {
			p.bIncrement()
		}
	case strings.Contains(line, "}"):
		for i := 0; i < strings.Count(line, "}"); i++ {
			p.bDecrement()
		}
	}
}

func (p *parserSrc) countParen(line string) {
	switch {
	case strings.Contains(line, "(") && strings.Contains(line, ")"):
		for i := 0; i < strings.Count(line, "("); i++ {
			p.pIncrement()
		}
		for i := 0; i < strings.Count(line, ")"); i++ {
			p.pDecrement()
		}
	case strings.Contains(line, "("):
		for i := 0; i < strings.Count(line, "("); i++ {
			p.pIncrement()
		}
	case strings.Contains(line, ")"):
		for i := 0; i < strings.Count(line, ")"); i++ {
			p.pDecrement()
		}
	}
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

func (p *parserSrc) parserImport(line string, iq chan<- importSpec) bool {
	var pat string
	if p.imFlag {
		pat = `\A[[:blank:]]*(\(?)([[:blank:]]*((.|\S+)[[:blank:]]+)?"([\S/]+)")?[[:blank:]]*(\)?)[[:blank:]]*\z`
	} else {
		pat = `\A[[:blank:]]*import[[:blank:]]*(\(?)([[:blank:]]*((.|\S+)[[:blank:]]+)?"([\S/]+)")?[[:blank:]]*(\)?)[[:blank:]]*\z`
	}
	re := regexp.MustCompile(pat)
	group := re.FindStringSubmatch(line)
	if len(group) != 7 {
		return false
	}
	if group[1] == "(" || group[1] == "" && group[2] == "" && group[3] == "" {
		p.imFlag = true
	}
	if group[5] != "" {
		// group[5] is imPath
		// group[4] is pkgName or ""
		p.putPackages(group[5], group[4], iq)
	}
	if group[6] == ")" {
		p.imFlag = false
	}
	return true
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

func (p *parserSrc) parserFuncSignature(line string) bool {

	funcName := `[[:blank:]]*func[[:blank:]]+(\((\w+[[:blank:]]+\*?\w+)\)[[:blank:]]*)?(\w+)[[:blank:]]*`
	params := `([\w_\*\[\],[:blank:]]+|[:blank:]*)`
	result := `\(([\w_\*\[\],[:blank:]]+)\)|([\w\*\[\][:blank:]]*)`
	pat := fmt.Sprintf(`\A%s\(%s\)[[:blank:]]*(%s)[[:blank:]]*{[[:blank:]]*\z`, funcName, params, result)
	re := regexp.MustCompile(pat)
	num := re.NumSubexp()
	groups := re.FindAllStringSubmatch(line, num)
	// group[1]: type (or groups[2] without paren)
	// group[3]: FuntionName
	// group[4]: Parameters
	// group[6]: result (multiple)
	// group[5]: result (single)
	if len(groups) == 0 {
		return false
	}
	for _, group := range groups {
		if group[3] == "main" {
			// func main
			p.mainFlag = true
			p.main = append(p.main, line)
			p.funcFlag = ""
		} else {
			// func other than main
			var result string
			if group[6] != "" {
				// multiple results
				result = group[6]
			} else if group[5] != "" {
				// single result
				result = group[5]
			} else {
				result = ""
			}
			p.funcFlag = group[3]
			p.funcDecls = append(p.funcDecls, funcDecl{group[3], signature{group[1], group[4], result}, []string{}})
		}
		p.countBlackets(line)
	}
	return true
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
			p.typeDecls[i].methSpecs = append(p.typeDecls[i].methSpecs, methodSpecs{group[3], signature{methType, group[4], result}})
		} else {
			p.typeDecls = append(p.typeDecls, typeDecl{group[3], group[4], []fieldDecl{}, []methodSpecs{}})
		}
	}
	if p.typeFlag == "paren" {
		p.countParen(line)
	} else if p.typeFlag == "struct" || p.typeFlag == "interface" {
		p.countBlackets(line)
	}
	return true
}

func (p *parserSrc) parserTypeSpec(line string) bool {
	if p.typeFlag == "paren" {
		p.countParen(line)
		if strings.Contains(line, ")") && p.paren == 0 {
			p.typeFlag = ""
			return true
		}
	} else if p.typeFlag == "struct" || p.typeFlag == "interface" {
		p.countBlackets(line)
		if strings.Contains(line, "}") && p.blackets == 0 {
			p.typeFlag = ""
			return true
		}
	}
	return false
}

func (p *parserSrc) parserMainBody(line string) bool {
	if p.mainFlag {
		p.main = append(p.main, line)
		p.countBlackets(line)
		if strings.Contains(line, "}") && p.blackets == 0 {
			// closing func main
			p.mainFlag = false
			return true
		}
	}
	return false
}

func (p *parserSrc) parserFuncBody(line string) bool {
	if p.funcFlag != "" {
		// func body
		p.appendBody(line)
		p.countBlackets(line)
		if strings.Contains(line, "}") && p.blackets == 0 {
			// closing func main
			p.funcFlag = ""
		}
	} else {
		// parse body
		return false
	}
	return true
}

func (p *parserSrc) parseLine(line string, iq chan<- importSpec) bool {
	// Ignore `package main', etc.
	if p.ignoreStatement(line) {
		return false
	}

	switch {
	case p.parserImport(line, iq):
		// import parser
	case p.parserTypeSpec(line):
		// type spec of struct parser
	case p.parserType(line):
		// type parser
	case p.parserFuncSignature(line):
		// func signature parser
	case p.parserMainBody(line):
		// func main body parser
		return true
	case p.parserFuncBody(line):
		// func body parser
	default:
		// parser body
		if !p.mainFlag {
			p.countBlackets(line)
			p.body = append(p.body, line)
		}
	}
	return false
}

func convertImport(pkgs []importSpec) []string {
	// convert packages list to "import" statement

	lines := []string{"import (\n"}
	for _, pkg := range pkgs {
		lines = append(lines, fmt.Sprintf("%s \"%s\"\n", pkg.pkgName, pkg.imPath))
	}
	lines = append(lines, ")\n")
	return lines
}

func (p *parserSrc) convertFuncDecls() []string {
	var lines []string
	for _, fun := range p.funcDecls {
		lines = append(lines, fmt.Sprintf("func %s %s(%s) (%s) {", fun.sig.functype, fun.name, fun.sig.params, fun.sig.result))
		for _, l := range fun.body {
			lines = append(lines, l)
		}
	}
	return lines
}

func (p *parserSrc) convertTypeDecls() []string {
	lines := []string{"type ("}
	for _, t := range p.typeDecls {
		if len(t.methSpecs) > 0 {
			lines = append(lines, fmt.Sprintf("%s %s {", t.typeID, t.typeName))
			for _, m := range t.methSpecs {
				if m.sig.result == "" {
					lines = append(lines, fmt.Sprintf("%s%s(%s)", m.sig.functype, m.name, m.sig.params))
				} else {
					lines = append(lines, fmt.Sprintf("%s%s(%s) %s", m.sig.functype, m.name, m.sig.params, m.sig.result))
				}
			}
			lines = append(lines, "}")
		} else if len(t.fieldDecls) > 0 {
			lines = append(lines, fmt.Sprintf("%s %s {", t.typeID, t.typeName))
			for _, f := range t.fieldDecls {
				if f.fieldType != "" {
					lines = append(lines, fmt.Sprintf("%s %s", f.idList, f.fieldType))
				} else {
					lines = append(lines, f.idList)
				}
			}
			lines = append(lines, "}")
		} else {
			lines = append(lines, fmt.Sprintf("%s %s", t.typeID, t.typeName))
		}
	}
	lines = append(lines, ")")
	return lines
}

func (p *parserSrc) mergeLines() []string {
	// merge "package", "import", "func", "func main".
	lines := []string{"package main\n"}
	lines = append(lines, convertImport(p.imPkgs)...)
	lines = append(lines, p.convertTypeDecls()...)
	lines = append(lines, p.convertFuncDecls()...)
	lines = append(lines, p.body...)
	lines = append(lines, p.main...)
	return lines
}

func (p *parserSrc) ignoreStatement(line string) bool {
	// ignore statement
	switch {
	case p.ignorePkgClause(line):
		return true
	}
	return false
}

func (p *parserSrc) ignorePkgClause(line string) bool {
	// ignore PackageClause
	var pat string
	if p.pkgFlag {
		pat = `\A([[:blank:]]*package)?[[:blank:]]*([[\pL\d_]+)[[:blank:]]*\z`
	} else {
		pat = `\A([[:blank:]]*package)([[:blank:]]+[\pL\d_]+)?[[:blank:]]*\z`
	}
	re := regexp.MustCompile(pat)
	group := re.FindStringSubmatch(line)
	if len(group) != 3 {
		return false
	}
	if group[1] == "" {
		// `"package"'
		p.pkgFlag = false
	}

	if group[2] == "" {
		// PackageName
		p.pkgFlag = true
	}
	return true
}
