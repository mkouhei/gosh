package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func reader() ([]byte, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(">>> ")
	text, err := reader.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	return text, nil
}

func initFile(codePath string) error {
	if _, err := os.Stat(codePath); err == nil {
		return nil
	}
	f, err := os.OpenFile(codePath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond)

	f.WriteString("package main\n")
	f.Sync()
	f.Close()

	return nil
}

func writeFile(codePath string, content string) error {
	f, err := os.OpenFile(codePath, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	time.Sleep(time.Microsecond)

	f.WriteString(content)
	f.Sync()
	f.Close()

	return nil
}

func shell() {
	p := fmt.Sprintf("%s/%s", bldDir(), tmpname)
	fmt.Println(p)
	if err := initFile(p); err != nil {
		fmt.Printf("[error] %v", err)
		return
	}
	for {
		text, err := reader()
		if err != nil {
			break
		}
		fmt.Print(string(text))
		if err := writeFile(p, string(text)); err != nil {
			fmt.Printf("[error] %v", err)
			break
		}
	}
}
