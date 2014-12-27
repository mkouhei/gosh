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
	"strings"
	"testing"
)

var (
	src = `package main
import (
"fmt"
o "os"
"bytes"
)
type hoge int
type foo []string
type (
bar string
baz int
spam struct {
name string
cnt int
lines []string
}
ham interface {
Write()
Read(b bytes.Buffer) bool
List(l []string) (int, bool)
}
eggs []string
)
type qux struct {
name string
cnt int
}
type quux interface {
Write()
Read(b bytes.Buffer) bool
}
func test0() bool {
f, err := o.Stat("/tmp")
if err != nil {
return false
}
return f.IsDir()
}
func test1() {
fmt.Println("helo")
}
func test1() {
fmt.Println("hello")
}
func test2(cnt int) {
cnt += ((cnt+1)*2-3)/4%5
cnt *= 4
cnt -= 3
cnt /= 2
cnt %= 5
fmt.Printf("%d\n", cnt)
}
func test3(cnt int) string {
return fmt.Sprintf("%d\n", cnt)
}
func test4(msg string, cnt int) string {
return fmt.Sprintf("%d: %s\n", cnt, msg)
}
func test5(msgs []string,cnt int) {
for i, l := range msgs {
fmt.Printf("%d: %s\n", i + cnt, l)
}
}
func test6(msg string, cnt int) (string, int) {
return fmt.Sprintf("%d: %s\n", cnt, msg), 1
}
func (f foo) test7() {
fmt.Println(f)
}
func (q *qux) test8(name string) {
fmt.Println(q.name == name)
fmt.Println(q.name != name)
q.name = name
fmt.Println(q.name)
}
func main() {
if !test0() {
fmt.Println(test0())
}
test1()
test2(2)
fmt.Println(test3(3))
fmt.Println(test4("hello", 4))
msgs := []string{"bye"}
test5(msgs, 5)
fmt.Println(test6("hello, again", 6))
f := foo{"bye"}
f.test7()
q := qux{"bye bye", 1}
q.test8("end")
}
`
	typeResult = `type (
hoge int
foo []string
bar string
baz int
spam struct {
name string
cnt int
lines []string
}
qux struct {
name string
cnt int
}
ham interface {
Write()
Read(b bytes.Buffer) bool
List(l []string) (int, bool)
}
eggs []string
quux interface {
Write()
Read(b bytes.Buffer) bool
}
)
`

	mainResult = `if ! test0(){fmt.Println(test0())
}
test1()
test2(2)
fmt.Println(test3(3))
fmt.Println(test4("hello",4))
msgs:=[]string{"bye"}
test5(msgs,5)
fmt.Println(test6("hello, again",6))
f:=foo{"bye"}
f.test7()
q:=qux{"bye bye",1}
q.test8("end")
`
)

func consumeChan(iq <-chan importSpec) {
	go func() {
		for {
			<-iq
		}
	}()
}

func TestParseImportFail(t *testing.T) {
	p := parserSrc{}
	p.imPkgs = append(p.imPkgs, importSpec{})
	iq := make(chan importSpec, 1)
	consumeChan(iq)

	lines := []string{"import fmt"}

	for _, l := range lines {
		p.parseLine([]byte(l), iq)
	}

	if len(compareImportSpecs(p.imPkgs, []importSpec{})) != 1 {
		t.Fatal("parse error: expected nil")
	}
}

func TestParseMultipleImport(t *testing.T) {
	p := parserSrc{}
	iq := make(chan importSpec, 4)

	lines := []string{`import "fmt"`,
		`import "io"`,
		`import (`,
		`str "strings"`,
		`"os"`,
		`)`,
	}

	for _, l := range lines {
		p.parseLine([]byte(l), iq)
	}

	el := []importSpec{
		importSpec{"fmt", ""},
		importSpec{"io", ""},
		importSpec{"strings", "str"},
		importSpec{"os", ""}}
	if len(compareImportSpecs(p.imPkgs, el)) != 0 {
		t.Fatalf("parse error: expected %v", el)
	}
}

func TestParseDuplicateImport(t *testing.T) {
	p := parserSrc{}
	iq := make(chan importSpec, 2)

	lines := []string{`import "fmt"`,
		`import "fmt"`,
	}

	for _, l := range lines {
		p.parseLine([]byte(l), iq)
	}
	el := []importSpec{
		importSpec{"fmt", ""}}
	if len(compareImportSpecs(p.imPkgs, el)) != 0 {
		t.Fatalf(`parse error: expected %v`, el)
	}
}

func TestParseLine(t *testing.T) {
	p := parserSrc{}
	iq := make(chan importSpec, 10)

	lines := strings.Split(src, "\n")

	import1 := []importSpec{
		importSpec{"fmt", ""},
		importSpec{"os", "o"},
		importSpec{"bytes", ""}}

	type1 := strings.Split(typeResult, "\n")
	main1 := strings.Split(mainResult, "\n")

	for _, l := range lines {
		p.parseLine([]byte(l), iq)
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

	if len(p.mergeLines()) != 85 {
		t.Fatal("parse error")
	}
	if p.braces != 0 {
		t.Fatalf("braces count error: %d", p.braces)
	}

}
