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
	importPath  string
	packageName string
}

type parser struct {
	packageFlag  bool
	importPkgs   []importSpec
	importFlag   bool
	body         []string
	mainFlag     bool
	mainBlackets int32
	main         []string
	continuous   bool
}

func (p *parser) increment() {
	atomic.AddInt32(&p.mainBlackets, 1)
}

func (p *parser) decrement() {
	atomic.AddInt32(&p.mainBlackets, -1)
}

func (p *parser) putPackages(importPath, packageName string, iq chan<- importSpec) {
	// put package to queue of `go get'
	if !searchPackage(importSpec{importPath, packageName}, p.importPkgs) {
		is := importSpec{importPath, packageName}
		p.importPkgs = append(p.importPkgs, is)
		iq <- is
	}
}

func (p *parser) parserImport(line string, iq chan<- importSpec) bool {
	var pat string
	if p.importFlag {
		pat = "\\A[[:blank:]]*(\\(?)([[:blank:]]*((.|\\S+)[[:blank:]]+)?\"([\\S/]+)\")?[[:blank:]]*(\\)?)[[:blank:]]*\\z"
	} else {
		pat = "\\Aimport[[:blank:]]*(\\(?)([[:blank:]]*((.|\\S+)[[:blank:]]+)?\"([\\S/]+)\")?[[:blank:]]*(\\)?)[[:blank:]]*\\z"
	}
	re, _ := regexp.Compile(pat)
	group := re.FindStringSubmatch(line)
	if len(group) != 7 {
		return false
	}
	if group[1] == "(" || group[1] == "" && group[2] == "" && group[3] == "" {
		p.importFlag = true
	}
	if group[5] != "" {
		// group[5] is importPath
		// group[4] is packageName or ""
		p.putPackages(group[5], group[4], iq)
	}
	if group[6] == ")" {
		p.importFlag = false
	}
	return true
}

func removeImportPackage(slice *[]importSpec, pkg importSpec) {
	s := *slice
	for i, item := range s {
		if item.importPath == pkg.importPath && item.packageName == pkg.packageName {
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

func (p *parser) parseLine(line string, iq chan<- importSpec) bool {
	// Ignore `package main', etc.
	if p.ignoreStatement(line) {
		return false
	}

	switch {
	case p.parserImport(line, iq):
		// import parser

	case strings.HasPrefix(line, "func "):
		// parser "func"
		switch {
		case strings.Contains(line, "main"):
			// func main
			p.mainFlag = true
			p.increment()
			p.main = append(p.main, line)
		default:
			// func other than main
			p.body = append(p.body, line)
		}
	case p.mainFlag:
		p.main = append(p.main, line)
		switch {
		case strings.Contains(line, "{"):
			p.increment()
		case strings.Contains(line, "}"):
			p.decrement()
			if p.mainBlackets == 0 {
				// closing func main
				p.mainFlag = false
				return true
			}
		}
	default:
		p.body = append(p.body, line)
	}
	return false
}

func convertImport(pkgs []importSpec) []string {
	// convert packages list to "import" statement

	imports := []string{"import (\n"}
	for _, pkg := range pkgs {
		imports = append(imports, fmt.Sprintf("%s \"%s\"\n", pkg.packageName, pkg.importPath))
	}
	imports = append(imports, ")\n")
	return imports
}

func (p *parser) mergeLines() []string {
	// merge "package", "import", "func", "func main".
	lines := []string{"package main\n"}
	lines = append(lines, convertImport(p.importPkgs)...)
	lines = append(lines, p.body...)
	lines = append(lines, p.main...)
	return lines
}

func (p *parser) ignoreStatement(line string) bool {
	// ignore statement
	switch {
	case p.ignorePkgClause(line):
		return true
	}
	return false
}

func (p *parser) ignorePkgClause(line string) bool {
	// ignore PackageClause
	var pat string
	if p.packageFlag {
		pat = "\\A([[:blank:]]*package)?[[:blank:]]*([[\\pL\\d_]+)[[:blank:]]*\\z"
	} else {
		pat = "\\A([[:blank:]]*package)([[:blank:]]+[\\pL\\d_]+)?[[:blank:]]*\\z"
	}
	re, _ := regexp.Compile(pat)
	group := re.FindStringSubmatch(line)
	if len(group) != 3 {
		return false
	}
	if group[1] == "" {
		// `"package"'
		p.packageFlag = false
	}

	if group[2] == "" {
		// PackageName
		p.packageFlag = true
	}
	return true
}
