package main

/*
   Copyright (C) 2014 Kouhei Maeda <mkouhei@palmtb.net>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

import (
	"testing"
)

func consumeChan(iq <-chan importSpec) {
	go func() {
		for {
			<-iq
		}
	}()
}

func TestParseImportFail(t *testing.T) {
	p := parser{}
	p.importPkgs = append(p.importPkgs, importSpec{})
	iq := make(chan importSpec, 1)
	consumeChan(iq)

	lines := []string{"import fmt"}

	for _, l := range lines {
		p.parseLine(l, iq)
	}

	if len(compareImportSpecs(p.importPkgs, []importSpec{})) != 1 {
		t.Fatal("parse error: expected nil")
	}
}

func TestParseMultipleImport(t *testing.T) {
	p := parser{}
	iq := make(chan importSpec, 4)

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

	el := []importSpec{
		importSpec{"fmt", ""},
		importSpec{"io", ""},
		importSpec{"strings", ""},
		importSpec{"os", ""}}
	if len(compareImportSpecs(p.importPkgs, el)) != 0 {
		t.Fatalf("parse error: expected %v", el)
	}
}

func TestParseDuplicateImport(t *testing.T) {
	p := parser{}
	iq := make(chan importSpec, 2)

	lines := []string{"import \"fmt\"",
		"import \"fmt\"",
	}

	for _, l := range lines {
		p.parseLine(l, iq)
	}
	el := []importSpec{
		importSpec{"fmt", ""}}
	if len(compareImportSpecs(p.importPkgs, el)) != 0 {
		t.Fatalf("parse error: expected %v", el)
	}
}

func TestParseLine(t *testing.T) {
	p := parser{}
	iq := make(chan importSpec, 10)

	lines := []string{"package main",
		"import (",
		"\"fmt\"",
		"\"os\"",
		")",
		"func test0() bool {",
		"f, err := os.Stat(\"/tmp\")",
		"if err != nil {",
		"return false",
		"}",
		"return f.IsDir()",
		"}",
		"func test1() {",
		"fmt.Println(\"hello\")",
		"}",
		"func test2(cnt int) {",
		"fmt.Printf(\"%d\n\", cnt)",
		"}",
		"func test3(cnt int) string {",
		"return fmt.Sprintf(\"%d\n\", cnt)",
		"}",
		"func test4(msg string, cnt int) string {",
		"return fmt.Sprintf(\"%d: %s\n\", cnt msg)",
		"}",
		"func test5(msg string,cnt int) string {",
		"return fmt.Sprintf(\"%d: %s\n\", cnt msg)",
		"}",
		"func test6(msg string, cnt int) (string, int) {",
		"return fmt.Sprintf(\"%d: %s\n\", cnt msg), 1",
		"}",
		"type foo string",
		"func (f foo) test7() {",
		"fmt.Println(f)",
		"}",
		"func main() {",
		"fmt.Println(test())",
		"}",
	}

	import1 := []importSpec{
		importSpec{"fmt", ""},
		importSpec{"os", ""}}

	body1 := []string{"type foo string"}

	main1 := []string{
		"func main() {",
		"fmt.Println(test())",
		"}",
	}

	for _, l := range lines {
		p.parseLine(l, iq)
	}

	if len(compareImportSpecs(p.importPkgs, import1)) != 0 {
		t.Fatal("parse error")
	}
	if len(compare(p.body, body1)) != 0 {
		t.Fatal("parse error")
	}

	if len(compare(p.main, main1)) != 0 {
		t.Fatal("parse error")
	}
	if len(p.mergeLines()) != 37 {
		t.Fatal("parse error")
	}

}
