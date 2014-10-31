package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func reader(in *os.File) (string, error) {
	if in == nil {
		in = os.Stdin
	}
	reader := bufio.NewReader(in)
	fmt.Print(">>> ")
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
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
	tmpDir := bldDir()

	l := lines{}
	p := fmt.Sprintf("%s/%s", tmpDir, tmpname)
	if err := initFile(p); err != nil {
		fmt.Printf("[error] %v", err)
		return
	}

	for {
		text, err := reader(nil)
		if err != nil {
			cleanDirs(tmpDir)
			break
		}
		l.parserImport(text)
		if err := writeFile(p, text); err != nil {
			fmt.Printf("[error] %v", err)
			break
		}
	}
}
