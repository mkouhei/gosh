package main

import (
	"testing"
)

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
