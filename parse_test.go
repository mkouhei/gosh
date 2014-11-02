package main

import (
	"testing"
)

func TestParserImport(t *testing.T) {
	expect_lines := []string{"fmt", "os", "io"}
	p := parser{}
	p.parserImport("foo")
	p.parserImport("")
	p.parserImport("\n")
	p.parserImport("import (")
	p.parserImport("\"fmt\"")
	p.parserImport("\"os\"")
	p.parserImport(")")
	p.parserImport("import \"io\"")
	if len(p.importPkgs) != 3 {
		t.Fatal("parse error")
	}

	if len(compare(expect_lines, p.importPkgs)) != 0 {
		t.Fatal("parse error")
	}
}
