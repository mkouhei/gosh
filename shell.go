package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func read(in *os.File) (string, error) {
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

func (e *env) initFile() error {
	if _, err := os.Stat(e.TmpPath); err == nil {
		return nil
	}
	f, err := os.OpenFile(e.TmpPath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond)

	f.WriteString("package main\n")
	f.Sync()
	f.Close()

	return nil
}

func (e *env) write(content string) error {
	f, err := os.OpenFile(e.TmpPath, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	time.Sleep(time.Microsecond)

	f.WriteString(content)
	f.Sync()
	f.Close()

	return nil
}

func (e env) WriteFile(lines []string) {
	for _, l := range lines {
		e.write(l)
	}
}

func (e *env) shell() bool {
	if err := e.initFile(); err != nil {
		e.logger("initFile", "", err)
		return false
	}
	go func() {
		if err := e.watch(); err != nil {
			e.logger("watch", "", err)
		}
	}()

	p := parser{[]string{}, false, []string{}, false, 0, []string{}, false}
	for {
		text, err := read(nil)
		if err != nil {
			cleanDirs(e.BldDir)
			return false
		}
		p.parseLine(text)
		e.logger("read", text, nil)
		if p.mainClosed {
			e.logger("mainClosed", "true", nil)
			break
		} else {
			e.logger("mainClosed", "false", nil)
		}
	}
	e.WriteFile(convertImport(p.importPkgs))
	e.WriteFile(p.body)
	e.WriteFile(p.main)
	return true
}
