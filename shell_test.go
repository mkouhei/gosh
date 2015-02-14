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
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	testSrc = `package main
import (
"fmt"
"os"
)
func main() {
fmt.Println("hello")
}
`

	testSrc2 = `import "fmt"

func main() {
     fmt.Println("hello")
}
`

	testSrc3 = `main() {
     fmt.Println("hello")
}
`

	expSrc = `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
)

func TestWrite(t *testing.T) {
	e := newEnv(false)
	imptCh := make(chan bool)
	imptQ := make(chan imptSpec, 10)
	for _, l := range strings.SplitAfter(testSrc, "\n") {
		e.parserSrc.parseLine([]byte(l), imptQ)
	}
	e.write(imptCh)
	time.Sleep(time.Microsecond)
	_, err := os.Stat(e.tmpPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGoImports(t *testing.T) {

	e := newEnv(true)
	fp, err := os.OpenFile(e.tmpPath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		t.Fatal(err)
	}
	fp.WriteString(testSrc)
	fp.Sync()
	fp.Close()

	time.Sleep(time.Microsecond)

	execCh := make(chan bool)
	e.goImports(execCh)

	time.Sleep(time.Microsecond)

	lines := []string{}
	if <-execCh {
		fp2, err := os.Open(e.tmpPath)
		if err != nil {
			t.Fatal(err)
		}
		s := bufio.NewScanner(fp2)
		for s.Scan() {
			lines = append(lines, s.Text())
		}
		fp2.Close()
	}

	expectLines := strings.Split(expSrc, "\n")
	if len(compare(lines, expectLines)) != 0 {
		t.Fatal("goimports error")
	}

}

func TestGoImportsFail(t *testing.T) {

	e := newEnv(true)
	fp, _ := os.OpenFile(e.tmpPath, os.O_WRONLY|os.O_CREATE, 0600)
	fp.WriteString(testSrc3)
	fp.Sync()
	fp.Close()

	time.Sleep(time.Microsecond)

	execCh := make(chan bool)
	e.goImports(execCh)

	time.Sleep(time.Microsecond)

	lines := []string{}
	if <-execCh {
		fp2, err := os.Open(e.tmpPath)
		if err != nil {
			t.Fatal(err)
		}
		s := bufio.NewScanner(fp2)
		for s.Scan() {
			lines = append(lines, s.Text())
		}
		fp2.Close()
	}

	expectLines := strings.Split(expSrc, "\n")
	if len(compare(lines, expectLines)) != 2 {
		t.Fatal("expects goimports error")
	}
}

func TestRead(t *testing.T) {
	e := newEnv(false)
	f, err := os.OpenFile("dummy_code", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	f.WriteString(testSrc2)

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
	imptQ := make(chan imptSpec, 1)
	imptQ <- imptSpec{"fmt", ""}
	e.goGet(imptQ)
	// Output:
	//
}

func ExampleGoRun() {
	e := newEnv(true)
	fp, err := os.OpenFile(e.tmpPath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(time.Microsecond)
	fp.WriteString(expSrc)
	fp.Sync()
	fp.Close()

	_, err = os.Stat(e.tmpPath)
	if err != nil {
		fmt.Println(err)
	}
	e.goRun()
	// Output:
	// hello

}

func ExampleGoRunFail() {
	e := newEnv(false)
	e.goRun()
	// Output:
	// [error] stat gosh_tmp.go: no such file or directory
}
