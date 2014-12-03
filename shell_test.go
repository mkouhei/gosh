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
	"bufio"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestGoImports(t *testing.T) {
	dummy := `package main
import (
"fmt"
"os"
)
func main() {
fmt.Println("hello")
}
`
	expectLines := []string{"package main",
		"",
		"import \"fmt\"",
		"",
		"func main() {",
		"\tfmt.Println(\"hello\")",
		"}",
	}

	e := newEnv(true)
	fp, err := os.OpenFile(e.TmpPath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		t.Fatal(err)
	}
	fp.WriteString(dummy)
	fp.Sync()
	fp.Close()

	time.Sleep(time.Microsecond)

	ec := make(chan bool)
	e.goImports(ec)

	time.Sleep(time.Microsecond)

	lines := []string{}
	if <-ec {
		fp2, err := os.Open(e.TmpPath)
		if err != nil {
			t.Fatal(err)
		}
		s := bufio.NewScanner(fp2)
		for s.Scan() {
			lines = append(lines, s.Text())
		}
		fp2.Close()
	}

	if len(compare(lines, expectLines)) != 0 {
		t.Fatal("goimports error")
	}

}

func TestRead(t *testing.T) {
	e := newEnv(false)
	f, err := os.OpenFile("dummy_code", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	text := `import "fmt"

func main() {
     fmt.Println("hello")
}
`
	f.WriteString(text)

	f, err = os.Open("dummy_code")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	e.shell(f)

	time.Sleep(time.Nanosecond)

	os.Remove("dummy_code")
}

func ExampleGoGet() {
	e := newEnv(false)
	iq := make(chan importSpec, 1)
	iq <- importSpec{"fmt", ""}
	e.goGet(iq)
	// Output:
	//
}

func ExampleGoRun() {
	dummy := `package main
import "fmt"
func main() {
fmt.Println("hello")
}
`
	e := newEnv(true)
	fp, err := os.OpenFile(e.TmpPath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(time.Microsecond)
	fp.WriteString(dummy)
	fp.Sync()
	fp.Close()

	_, err = os.Stat(e.TmpPath)
	if err != nil {
		fmt.Println(err)
	}
	e.goRun()
	// Output:
	// hello

}

func TestRemoveImport(t *testing.T) {
	e := newEnv(false)
	pkgs := []importSpec{
		importSpec{"fmt", ""},
		importSpec{"os", ""},
		importSpec{"hoge", ""},
		importSpec{"io", ""}}
	pkgs2 := []importSpec{
		importSpec{"fmt", ""},
		importSpec{"os", ""},
		importSpec{"io", ""}}
	e.parser.imPkgs = pkgs

	e.removeImport("dummy message", importSpec{"hoge", ""})
	if len(compareImportSpecs(e.parser.imPkgs, pkgs)) != 0 {
		t.Fatal("fail filtering")
	}

	e.removeImport("package moge: unrecognized import path \"moge\"", importSpec{"hoge", ""})
	if len(compareImportSpecs(e.parser.imPkgs, pkgs)) != 0 {
		t.Fatal("fail filtering")
	}

	e.removeImport("package hoge: unrecognized import path \"hoge\"", importSpec{"hoge", ""})
	if len(compareImportSpecs(e.parser.imPkgs, pkgs2)) != 0 {
		t.Fatal("fail remove package")
	}
}
