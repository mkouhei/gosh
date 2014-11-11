package main

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
	return string(group[1])
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
			p.putPackages(pkg, iq)
		}
	} else if p.importFlag {
		if strings.HasPrefix(line, ")") {
			p.importFlag = false
		} else {
			r := strings.NewReader(line)
			if r.Len() > 0 {
				pkg := pkgName(line)
				p.putPackages(pkg, iq)
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
