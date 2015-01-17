package main

/*
   Copyright (C) 2014,2015 Kouhei Maeda <mkouhei@palmtb.net>

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
"net/http"
"github.com/bitly/go-simplejson"
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
func test1() []string {
return []string{"hello"}
}
func test2(cnt int) {
cnt << 1
cnt += ((cnt+1)*2-3)/4%5
cnt *= 4
cnt -= 3
cnt /= 2
cnt %= 5
cnt++
cnt >> 1
fmt.Printf("%d\n", cnt)
}
func test3(cnt int) string {
var msg string
switch {
case cnt == 0 || cnt == 1:
msg = "0,1"
case cnt > 1 && cnt < 4:
msg = "2,3"
case cnt >= 4 && cnt <= 10:
msg = "4-10"
default:
msg = "none"
}
return msg
}
func test4(msg string, cnt int) string {
return fmt.Sprintf("%d: %s\n", cnt, msg)
}
func test5(msgs []string,cnt int) {
cnt--
for i, l := range msgs {
fmt.Printf("%d: %s\n", i + cnt, l)
}
}
func test6(msg string, cnt int) (string, int) {
return fmt.Sprintf("%d: %s\n", cnt, msg), 1
}
func (f foo) test7() (*qux, *int) {
q := &qux{}
var c int
if len(f) == 0 {
c = -1
return nil, &c
}
for i, str := range f {
q.cnt = i
q.name = str
if i > 1 {
break
}
}
c = 0
return q, &c
}
func (q *qux) test8(name string) {
fmt.Println(q.name == name)
fmt.Println(q.name != name)
q.name = name
fmt.Println(q.name)
}
func test9(url string) *simplejson.Json {
resp, _ := http.Get(url)
defer resp.Body.Close()
js, _ := simplejson.NewFromReader(resp.Body)
return js
}
func main() {
if !test0() {
fmt.Println(test0())
}
fmt.Println(test1())
test2(2)
fmt.Println(test3(3))
fmt.Println(test4("hello", 4))
msgs := []string{"bye"}
test5(msgs, 5)
fmt.Println(test6("hello, again", 6))
f := foo{"bye"}
fmt.Println(f.test7())
q := qux{"bye bye", 1}
q.test8("end")
test9("http://example.org/dummy.json")
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
ham interface {
Write()
Read(b bytes.Buffer) bool
List(l []string) (int, bool)
}
eggs []string
qux struct {
name string
cnt int
}
quux interface {
Write()
Read(b bytes.Buffer) bool
}
)`

	funcResult = `func test0() bool {
f, err := o.Stat("/tmp")
if err != nil{
return false
}
return f.IsDir()
}
func test1() []string {
return []string{"hello"}
}
func test2(cnt int) {
cnt << 1
cnt +=((cnt+1)*2-3)/4%5
cnt *= 4
cnt -= 3
cnt /= 2
cnt %= 5
cnt++
cnt >> 1
fmt.Printf("%d\n", cnt)
}
func test3(cnt int) string {
var msg string
switch{case cnt == 0 || cnt == 1:
msg = "0,1"
case cnt > 1 && cnt < 4:
msg = "2,3"
case cnt >= 4 && cnt <= 10:
msg = "4-10"
default:
msg = "none"
}
return msg
}
func test4(msg string, cnt int) string {
return fmt.Sprintf("%d: %s\n", cnt, msg)
}
func test5(msgs []string, cnt int) {
cnt--
for i, l := range msgs{fmt.Printf("%d: %s\n", i+cnt, l)
}
}
func test6(msg string, cnt int) (string, int) {
return fmt.Sprintf("%d: %s\n", cnt, msg), 1
}
func (f foo)test7() (*qux, *int) {
q :=&qux{}
var c int
if len(f)== 0{c =-1
return nil, &c
}
for i, str := range f{q.cnt = i
q.name = str
if i > 1{
break
}
}
c = 0
return q, &c
}
func (q *qux)test8(name string) {
fmt.Println(q.name == name)
fmt.Println(q.name != name)
q.name = name
fmt.Println(q.name)
}
func test9(url string) *simplejson.Json {
resp, _ := http.Get(url)
defer resp.Body.Close()
js, _ := simplejson.NewFromReader(resp.Body)
return js
}`

	mainResult = `if ! test0(){
fmt.Println(test0())
}
fmt.Println(test1())
test2(2)
fmt.Println(test3(3))
fmt.Println(test4("hello", 4))
msgs := []string{"bye"}
test5(msgs, 5)
fmt.Println(test6("hello, again", 6))
f := foo{"bye"}
fmt.Println(f.test7())
q := qux{"bye bye", 1}
q.test8("end")
test9("http://example.org/dummy.json")`

	mainOmit = `if true {
fmt.Println("hello")
}
`

	mainOmitResult = `if true{fmt.Println("hello")
}`
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
		`import "bytes"`,
		`import "fmt"`,
		`import "bytes"`,
		`import "fmt"`,
	}

	for _, l := range lines {
		p.parseLine([]byte(l), iq)
	}
	el := []importSpec{
		importSpec{"fmt", ""},
		importSpec{"bytes", ""}}
	if len(compareImportSpecs(p.imPkgs, el)) != 0 {
		t.Fatalf(`parse error: expected %v`, el)
	}
}

func TestRemovePrintStmt(t *testing.T) {
	s := []string{"i := 1",
		"fmt.Println(i)"}
	es := []string{"i := 1"}
	removePrintStmt(&s)
	if len(s) != 1 {
		t.Fatalf("remove error: expected %v", es)
	}
	s2 := []string{"i := 1",
		"if i == 1 {",
		`fmt.Println("helo")`,
		"i++",
		"fmt.Println(i)",
		"}"}
	es2 := []string{"i := 1",
		"if i == 1 {",
		"i++",
		"}"}
	removePrintStmt(&s2)
	if len(s2) != 4 {
		t.Fatalf("remove error: expected %v, got %v", es2, s2)
	}
}

func TestParseLine(t *testing.T) {
	p := parserSrc{}
	iq := make(chan importSpec, 10)

	lines := strings.Split(src, "\n")

	import1 := []importSpec{
		importSpec{"fmt", ""},
		importSpec{"os", "o"},
		importSpec{"bytes", ""},
		importSpec{"net/http", ""},
		importSpec{"github.com/bitly/go-simplejson", ""}}

	type1 := strings.Split(typeResult, "\n")
	func1 := strings.Split(funcResult, "\n")
	main1 := strings.Split(mainResult, "\n")

	for _, l := range lines {
		p.parseLine([]byte(l), iq)
	}

	if len(compareImportSpecs(p.imPkgs, import1)) != 0 {
		t.Fatal("parse import packages error")
	}

	if len(compare(p.body, []string{})) != 0 {
		t.Fatal("parse body error")
	}

	if len(compare(p.convertTypeDecls(), type1)) != 0 {
		t.Fatal("parse type decls error")
	}

	if len(compare(p.convertFuncDecls(), func1)) != 0 {
		t.Fatal("parse func decls error")
	}

	if len(compare(p.main, main1)) != 0 {
		t.Fatal("parse main func error")
	}

	if len(p.mergeLines()) != 122 {
		t.Fatal("parse error")
	}
	if p.braces != 0 {
		t.Fatalf("braces count error: %d", p.braces)
	}

}

func TestOmit(t *testing.T) {
	p := parserSrc{}
	iq := make(chan importSpec, 1)
	lines := strings.Split(mainOmit, "\n")

	for _, l := range lines {
		p.parseLine([]byte(l), iq)
	}

	el := strings.Split(mainOmitResult, "\n")
	if len(compare(p.main, el)) != 0 {
		t.Fatal("parse omitted main func error")
	}
}
