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
	mainClosed   bool
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

func (p *parser) parseLine(line string) {

	if strings.HasPrefix(line, "import ") {
		if strings.Contains(line, "(") {
			p.importFlag = true
		} else {
			pkg := pkgName(strings.Split(line, " ")[1])
			goGet(pkg)
			p.importPkgs = append(p.importPkgs, pkg)
		}
	} else if p.importFlag {
		if strings.HasPrefix(line, ")") {
			p.importFlag = false
		} else {
			r := strings.NewReader(line)
			if r.Len() > 0 {
				goGet(pkgName(line))
				p.importPkgs = append(p.importPkgs, pkgName(line))
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
				p.mainClosed = true
			}
		}
	} else {
		p.body = append(p.body, line)
	}
}

func convertImport(pkgs []string) []string {

	imports := []string{"import (\n"}
	for _, pkg := range pkgs {
		imports = append(imports, fmt.Sprintf("\"%s\"\n", pkg))
	}
	imports = append(imports, ")\n")
	return imports
}
