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

func writeFile(codePath string, content string) error {
	f, err := os.OpenFile(codePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
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
	for {
		text, err := reader()
		if err != nil {
			break
		}
		fmt.Print(string(text))
		if err := writeFile(tmpname, string(text)); err != nil {
			fmt.Printf("[error] %v", err)
			break
		}
	}
}
