package main

import (
	"testing"
)

func consumeChan(iq <-chan string) {
	go func() {
		for {
			<-iq
		}
	}()
}

func TestParseImportFail(t *testing.T) {
	p := parser{}
	p.importPkgs = append(p.importPkgs, "")
	iq := make(chan string, 1)
	consumeChan(iq)

	lines := []string{"import fmt"}

	for _, l := range lines {
		p.parseLine(l, iq)
	}

	if len(compare(p.importPkgs, []string{})) != 1 {
		t.Fatal("parse error: expected nil")
	}
}

func TestParseMultipleImport(t *testing.T) {
	p := parser{}
	iq := make(chan string, 4)

	lines := []string{"import \"fmt\"",
		"import \"io\"",
		"import (",
		"\"strings\"",
		"\"os\"",
		")",
	}

	for _, l := range lines {
		p.parseLine(l, iq)
	}

	el := []string{"fmt", "io", "strings", "os"}
	if len(compare(p.importPkgs, el)) != 0 {
		t.Fatalf("parse error: expected %v", el)
	}
}

func TestParseDuplicateImport(t *testing.T) {
	p := parser{}
	iq := make(chan string, 1)

	lines := []string{"import \"fmt\"",
		"import \"fmt\"",
	}

	for _, l := range lines {
		p.parseLine(l, iq)
	}

	el := []string{"fmt"}
	if len(compare(p.importPkgs, el)) != 0 {
		t.Fatalf("parse error: expected %v", el)
	}
}

func TestParseLine(t *testing.T) {
	p := parser{}
	iq := make(chan string, 10)

	lines := []string{"import (",
		"\"fmt\"",
		"\"os\"",
		")",
		"func test() bool {",
		"f, err := os.Stat(\"/tmp\")",
		"if err != nil {",
		"return false",
		"}",
		"return f.IsDir()",
		"}",
		"func main() {",
		"fmt.Println(test())",
		"}",
	}

	import1 := []string{"fmt", "os"}
	body1 := []string{"func test() bool {",
		"f, err := os.Stat(\"/tmp\")",
		"if err != nil {",
		"return false",
		"}",
		"return f.IsDir()",
		"}",
	}

	main1 := []string{
		"func main() {",
		"fmt.Println(test())",
		"}",
	}

	for _, l := range lines {
		p.parseLine(l, iq)
	}

	if len(compare(p.importPkgs, import1)) != 0 {
		t.Fatal("parse error")
	}
	if len(compare(p.body, body1)) != 0 {
		t.Fatal("parse error")
	}
	if len(compare(p.main, main1)) != 0 {
		t.Fatal("parse error")
	}
	if len(p.convertLines()) != 15 {
		t.Fatal("parse error")
	}
}
