package main

import (
	"bufio"
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

func TestRead(t *testing.T) {
	e := NewEnv(false)
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
	e := NewEnv(false)
	iq := make(chan string, 1)
	iq <- "foo"
	e.goGet(iq)
	// Output:
	//
}
