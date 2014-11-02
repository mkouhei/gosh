package main

import (
	"regexp"
	"strings"
)

type parser struct {
	importPkgs []string
	importFlag bool
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
	}
}
