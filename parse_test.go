package main

import (
	"testing"
)

func TestParseLine(t *testing.T) {
	expect_lines := []string{"fmt", "os", "io"}
	p := parser{}
	p.parseLine("foo")
	p.parseLine("")
	p.parseLine("\n")
	p.parseLine("import (")
	p.parseLine("\"fmt\"")
	p.parseLine("\"os\"")
	p.parseLine(")")
	p.parseLine("import \"io\"")
	if len(p.importPkgs) != 3 {
		t.Fatal("parse error")
	}

	if len(compare(expect_lines, p.importPkgs)) != 0 {
		t.Fatal("parse error")
	}
}
