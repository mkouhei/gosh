package main

import (
	"regexp"
	"strings"
)

type lines struct {
	Import     []string
	importFlag bool
}

func pkgName(p string) string {
	re, _ := regexp.Compile("\"([\\S\\s/\\\\]+)\"")
	group := re.FindStringSubmatch(p)
	return string(group[1])
}

func (l *lines) parserImport(line string) {

	if strings.HasPrefix(line, "import ") {
		if strings.Contains(line, "(") {
			l.importFlag = true
		} else {
			l.Import = append(l.Import, pkgName(strings.Split(line, " ")[1]))
		}
	} else if l.importFlag {
		if strings.HasPrefix(line, ")") {
			l.importFlag = false
		} else {
			r := strings.NewReader(line)
			if r.Len() > 0 {
				l.Import = append(l.Import, pkgName(line))
			}
		}
	}
}
