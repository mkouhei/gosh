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
	re, _ := regexp.Compile("\"([\\S\\s/\\\\]+)\"")
	group := re.FindStringSubmatch(p)
	if len(group) != 0 {
		return string(group[1])
	}
	return ""
}

func (p *parser) putPackages(pkg string, iq chan<- string) {
	if !searchString(pkg, p.importPkgs) {
		p.importPkgs = append(p.importPkgs, pkg)
		iq <- pkg
	}
}

func (p *parser) parseLine(line string, iq chan<- string) bool {

	if strings.HasPrefix(line, "import ") {
		if strings.Contains(line, "(") {
			p.importFlag = true
		} else {
			pkg := pkgName(strings.Split(line, " ")[1])
			if pkg != "" {
				p.putPackages(pkg, iq)
			}
		}
	} else if p.importFlag {
		if strings.HasPrefix(line, ")") {
			p.importFlag = false
		} else {
			r := strings.NewReader(line)
			if r.Len() > 0 {
				pkg := pkgName(line)
				if pkg != "" {
					p.putPackages(pkg, iq)
				}
			}
		}
	} else if strings.HasPrefix(line, "func ") {
		if strings.Contains(line, "main") {
			// func main
			p.mainFlag = true
			p.increment()
			p.main = append(p.main, line)
		} else {
			// func other than main
			p.body = append(p.body, line)
		}
	} else if p.mainFlag {
		p.main = append(p.main, line)
		if strings.Contains(line, "{") {
			p.increment()
		} else if strings.Contains(line, "}") {
			p.decrement()
			if p.mainBlackets == 0 {
				// closing func main
				p.mainFlag = false
				return true
			}
		}
	} else {
		p.body = append(p.body, line)
	}
	return false
}

func convertImport(pkgs []string) []string {

	imports := []string{"import (\n"}
	for _, pkg := range pkgs {
		imports = append(imports, fmt.Sprintf("\"%s\"\n", pkg))
	}
	imports = append(imports, ")\n")
	return imports
}

func (p *parser) convertLines() []string {
	lines := []string{"package main\n"}
	lines = append(lines, convertImport(p.importPkgs)...)
	lines = append(lines, p.body...)
	lines = append(lines, p.main...)
	return lines
}
