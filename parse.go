package main

import (
	"strings"
)

type lines struct {
	Import     []string
	importFlag bool
}

func (l *lines) parserImport(line string) {

	if strings.HasPrefix(line, "import ") {
		if strings.Contains(line, "(") {
			l.importFlag = true
		} else {
			l.Import = append(l.Import, strings.Split(line, " ")[1])
		}
	} else if l.importFlag {
		if strings.HasPrefix(line, ")") {
			l.importFlag = false
		} else {
			r := strings.NewReader(line)
			if r.Len() > 0 {
				l.Import = append(l.Import, line)
			}
		}
	}
}
