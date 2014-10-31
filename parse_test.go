package main

import (
	"testing"
)

func compare(A, B []string) []string {
	m := make(map[string]int)
	for _, b := range B {
		m[b]++
	}
	var ret []string
	for _, a := range A {
		if m[a] > 0 {
			m[a]--
			continue
		}
		ret = append(ret, a)
	}
	return ret
}

func TestParserImport(t *testing.T) {
	expect_lines := []string{"fmt", "os", "io"}
	l := lines{}
	l.parserImport("foo")
	l.parserImport("")
	l.parserImport("\n")
	l.parserImport("import (")
	l.parserImport("\"fmt\"")
	l.parserImport("\"os\"")
	l.parserImport(")")
	l.parserImport("import \"io\"")
	if len(l.Import) != 3 {
		t.Fatal("parse error")
	}

	if len(compare(expect_lines, l.Import)) != 0 {
		t.Fatal("parse error")
	}
}
