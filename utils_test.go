package main

import (
	"bufio"
	"os"
	"testing"
	"time"
)

func TestRunCmd(t *testing.T) {
	cmd := "foo"
	args := []string{}
	if err := runCmd(cmd, args...); err == nil {
		t.Fatal("want: <fail>")
	}
	cmd = "true"
	if err := runCmd(cmd, args...); err != nil {
		t.Fatal(err)
	}
}

func ExampleGoGet() {
	e := NewEnv(false)
	e.goGet("foo")
	// Output:
	//
}

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
		"import \"fmt\"",
		"",
		"func main() {",
		"\tfmt.Println(\"hello\")",
		"}",
	}

	fp, _ := os.OpenFile("dummy_code", os.O_WRONLY|os.O_CREATE, 0600)
	defer fp.Close()
	fp.WriteString(dummy)

	ec := make(chan bool)

	e := NewEnv(false)
	e.TmpPath = "dummy_code"
	e.goImports(ec)

	time.Sleep(time.Microsecond)

	lines := []string{}
	go func() {
		<-ec
		fp2, err := os.Open("dummy_code")
		if err != nil {
			t.Fatal(err)
		}
		defer fp2.Close()
		s := bufio.NewScanner(fp2)
		for s.Scan() {
			lines = append(lines, s.Text())
		}

	}()

	if len(compare(lines, expectLines)) != 0 {
		t.Fatal("goimports error")
	}

}

func TestBldDirAndCleanDir(t *testing.T) {
	d := bldDir()
	f, err := os.Stat(d)
	if err != nil {
		t.Fatal(err)
	}
	if !f.IsDir() {
		t.Fatalf("expecting directory: %s", d)
	}
	if err := cleanDir(d); err != nil {
		t.Fatal(err)
	}
}

func TestSearchString(t *testing.T) {
	list := []string{"foo", "bar", "baz"}
	if !searchString("foo", list) {
		t.Fatal("expecting true")
	}
	if searchString("hoge", list) {
		t.Fatal("expecting false")
	}
}
