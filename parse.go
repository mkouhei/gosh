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

type parser struct {
	importPkgs   []string
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

func pkgName(p string) string {
	// extract 'foo' from '"foo"'
	re, _ := regexp.Compile("\"([\\S\\s/\\\\]+)\"")
	group := re.FindStringSubmatch(p)
	if len(group) != 0 {
		return string(group[1])
	}
	return ""
}

func (p *parser) putPackages(pkg string, iq chan<- string) {
	// put package to queue of `go get'
	if !searchString(pkg, p.importPkgs) {
		p.importPkgs = append(p.importPkgs, pkg)
		iq <- pkg
	}
}

func (p *parser) parseLine(line string, iq chan<- string) bool {
	// Ignore `package main', etc.
	if ignoreStatement(line) {
		return false
	}

	switch {
	case strings.HasPrefix(line, "import "):
		// parser "import"
		switch {
		case strings.Contains(line, "("):
			p.importFlag = true
		default:
			if pkg := pkgName(strings.Split(line, " ")[1]); pkg != "" {
				p.putPackages(pkg, iq)
			}
		}

	case p.importFlag:
		switch {
		case strings.HasPrefix(line, ")"):
			p.importFlag = false
		default:
			if strings.NewReader(line).Len() > 0 {
				if pkg := pkgName(line); pkg != "" {
					p.putPackages(pkg, iq)
				}
			}
		}

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

func convertImport(pkgs []string) []string {
	// convert packages list to "import" statement

	imports := []string{"import (\n"}
	for _, pkg := range pkgs {
		imports = append(imports, fmt.Sprintf("\"%s\"\n", pkg))
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

func ignoreStatement(line string) bool {
	// ignore statement
	switch {
	case strings.HasPrefix(line, "package "):
		return true
	}
	return false
}
