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
	p.imPkgs = append(p.imPkgs, importSpec{})
	iq := make(chan importSpec, 1)
	consumeChan(iq)

	lines := []string{"import fmt"}

	for _, l := range lines {
		p.parseLine(l, iq)
	}

	if len(compareImportSpecs(p.imPkgs, []importSpec{})) != 1 {
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
	if len(compareImportSpecs(p.imPkgs, el)) != 0 {
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
	if len(compareImportSpecs(p.imPkgs, el)) != 0 {
		t.Fatalf("parse error: expected %v", el)
	}
}

func TestParserType(t *testing.T) {
	p := parser{false, []importSpec{}, false, []funcDecl{}, "", 0, 0, []typeDecl{}, "", []string{}, false, []string{}}

	line := "type foo bool"
	if !p.parserType(line) || p.typeDecls[0].typeID != "foo" || p.typeDecls[0].typeName != "bool" {
		t.Fatalf(`parser error: %s`, line)
	}

	line = "type ("
	if !p.parserType(line) || p.typeFlag != "paren" || len(p.typeDecls) != 1 {
		t.Fatalf(`parser error: %s`, line)
	}

	line = "type bar int"
	if p.parserType(line) || p.typeFlag != "paren" || len(p.typeDecls) != 1 {
		t.Fatalf(`parser error: %s`, line)
	}

	line = "bar int"
	if !p.parserType(line) || p.typeFlag != "paren" || len(p.typeDecls) != 2 || p.typeDecls[1].typeID != "bar" || p.typeDecls[1].typeName != "int" {
		t.Fatalf(`parser error: %s`, line)
	}

	line = ")"
	if p.parserType(line) {
		t.Fatalf(`parser error: %s`, line)
	}
	if !p.parserTypeSpec(line) || p.typeFlag != "" {
		t.Fatalf(`parser error: %s`, line)
	}

	line = "type baz struct {"
	if !p.parserType(line) || p.typeFlag != "struct" || len(p.typeDecls) != 3 || p.typeDecls[2].typeID != "baz" || p.typeDecls[2].typeName != "struct" || len(p.typeDecls[2].fieldDecls) != 0 {
		t.Fatalf(`parser error: %s`, line)
	}
	line = "key string"
	if !p.parserType(line) || p.typeFlag != "struct" || len(p.typeDecls[2].fieldDecls) != 1 {
		t.Fatalf(`parser error: %s`, line)
	}
	line = "value string"
	if !p.parserType(line) || p.typeFlag != "struct" || len(p.typeDecls[2].fieldDecls) != 2 {
		t.Fatalf(`parser error: %s`, line)
	}
	line = "}"
	if p.parserType(line) || p.typeFlag != "struct" {
		t.Fatalf(`parser error: %s`, line)
	}
	if !p.parserTypeSpec(line) || p.typeFlag != "" {
		t.Fatalf(`parser error: %s`, line)
	}

	line = "type qux interface {"
	if !p.parserType(line) || p.typeFlag != "interface" || len(p.typeDecls) != 4 || p.typeDecls[3].typeID != "qux" || p.typeDecls[3].typeName != "interface" || len(p.typeDecls[3].fieldDecls) != 0 {
		t.Fatalf(`parser error: %s`, line)
	}
	line = "Read()"
	if !p.parserType(line) || p.typeFlag != "interface" || len(p.typeDecls[3].methSpecs) != 1 {
		t.Fatalf(`parser error: %s`, line)
	}
	line = "Write(b buffer) bool"
	if !p.parserType(line) || p.typeFlag != "interface" || len(p.typeDecls[3].methSpecs) != 2 {
		t.Fatalf(`parser error: %s`, line)
	}
	line = "}"
	if p.parserType(line) || p.typeFlag != "interface" {
		t.Fatalf(`parser error: %s`, line)
	}
	if !p.parserTypeSpec(line) || p.typeFlag != "" {
		t.Fatalf(`parser error: %s`, line)
	}

}

func TestParserFuncSignature(t *testing.T) {
	p := parser{false, []importSpec{}, false, []funcDecl{}, "", 0, 0, []typeDecl{}, "", []string{}, false, []string{}}
	line := "func main() {"
	if !p.parserFuncSignature(line) || !p.mainFlag {
		t.Fatalf(`parser error: "%s"`, line)
	}
	p.mainFlag = false
	line = "func main(){"
	if !p.parserFuncSignature(line) || !p.mainFlag {
		t.Fatalf(`parser error: "%s"`, line)
	}
	p.mainFlag = false
	line = "func main (){"
	if !p.parserFuncSignature(line) || !p.mainFlag {
		t.Fatalf(`parser error: "%s"`, line)
	}
	p.mainFlag = false
	line = "func main () {"
	if !p.parserFuncSignature(line) || !p.mainFlag {
		t.Fatalf(`parser error: "%s"`, line)
	}
	p.mainFlag = false
	line = "func main () { "
	if !p.parserFuncSignature(line) || !p.mainFlag {
		t.Fatalf(`parser error: "%s"`, line)
	}
	p.mainFlag = false
	line = " func main () { "
	if !p.parserFuncSignature(line) || !p.mainFlag {
		t.Fatalf(`parser error: "%s"`, line)
	}
	p.mainFlag = false
	line = "func main ()"
	if p.parserFuncSignature(line) || p.mainFlag {
		t.Fatalf(`parser error: "%s"`, line)
	}

	p.mainFlag = false
	line = "func foo() {"
	if !p.parserFuncSignature(line) || p.mainFlag {
		t.Fatalf(`parser error: %s`, line)
	}

	line = "func foo(bar string) {"
	if !p.parserFuncSignature(line) || p.mainFlag {
		t.Fatalf(`parser error: %s`, line)
	}

	line = "func foo(bar, baz  string) {"
	if !p.parserFuncSignature(line) || p.mainFlag {
		t.Fatalf(`parser error: %s`, line)
	}

	line = "func foo(bar, baz  string) boo{"
	if !p.parserFuncSignature(line) || p.mainFlag {
		t.Fatalf(`parser error: %s`, line)
	}

	line = "func foo(bar, baz  string) (boo, int){"
	if !p.parserFuncSignature(line) || p.mainFlag {
		t.Fatalf(`parser error: %s`, line)
	}

	line = "func (q *qux) foo(bar, baz  string) (boo, int){"
	if !p.parserFuncSignature(line) || p.mainFlag {
		t.Fatalf(`parser error: %s`, line)
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
		"type foo []string",
		"type (",
		"bar string",
		"baz int",
		")",
		"type qux struct {",
		"name string",
		"cnt int",
		"}",
		"type quux interface {",
		"Write()",
		"Read(b buffer) bool",
		"}",
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
		"return fmt.Sprintf(\"%d: %s\n\", cnt, msg)",
		"}",
		"func test5(msg string,cnt int) string {",
		"return fmt.Sprintf(\"%d: %s\n\", cnt, msg)",
		"}",
		"func test6(msg string, cnt int) (string, int) {",
		"return fmt.Sprintf(\"%d: %s\n\", cnt, msg), 1",
		"}",
		"func (f foo) test7() {",
		"fmt.Println(f)",
		"}",
		"func (q *qux) test8() {",
		"fmt.Println(q.name)",
		"}",
		"func main() {",
		"fmt.Println(test0())",
		"test1()",
		"test2(2)",
		"fmt.Println(test3(3))",
		"fmt.Println(test4(\"hello\", 4))",
		"fmt.Println(test5(\"bye\", 5))",
		"fmt.Println(test6(\"hello, again\", 6))",
		"f := foo{\"bye\"}",
		"f.test7()",
		"q := qux{\"bye bye\", 1}",
		"q.test8()",
		"}",
	}

	import1 := []importSpec{
		importSpec{"fmt", ""},
		importSpec{"os", ""}}

	type1 := []string{"type (",
		"foo []string",
		"bar string",
		"baz int",
		"qux struct {",
		"name string",
		"cnt int",
		"}",
		"quux interface {",
		"Write()",
		"Read(b buffer) bool",
		"}",
		")"}

	main1 := []string{
		"func main() {",
		"fmt.Println(test0())",
		"test1()",
		"test2(2)",
		"fmt.Println(test3(3))",
		"fmt.Println(test4(\"hello\", 4))",
		"fmt.Println(test5(\"bye\", 5))",
		"fmt.Println(test6(\"hello, again\", 6))",
		"f := foo{\"bye\"}",
		"f.test7()",
		"q := qux{\"bye bye\", 1}",
		"q.test8()",
		"}",
	}

	for _, l := range lines {
		p.parseLine(l, iq)
	}

	if len(compareImportSpecs(p.imPkgs, import1)) != 0 {
		t.Fatal("parse error")
	}

	if len(compare(p.body, []string{})) != 0 {
		t.Fatal("parse error")
	}

	if len(compare(p.convertTypeDecls(), type1)) != 0 {
		t.Fatal("parse error")
	}

	if len(compare(p.main, main1)) != 0 {
		t.Fatal("parse error")
	}

	if len(p.mergeLines()) != 62 {
		t.Fatal("parse error")
	}
	if p.blackets != 0 {
		t.Fatalf("blacket count error: %d", p.blackets)
	}

}
