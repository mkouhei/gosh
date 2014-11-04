package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
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

func (e *env) resetFile() error {
	if err := ioutil.WriteFile(e.TmpPath, []byte{}, 0600); err != nil {
		return err
	}
	f, err := os.OpenFile(e.TmpPath, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	time.Sleep(time.Microsecond)

	f.WriteString("package main\n")
	f.Sync()
	f.Close()

	return nil
}

func (e *env) write(lines []string) error {
	f, err := os.OpenFile(e.TmpPath, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	for _, l := range lines {
		f.WriteString(l)
	}
	f.Sync()
	f.Close()
	return nil
}

func (e *env) shell() error {
	if err := e.initFile(); err != nil {
		e.logger("initFile", "", err)
		e.parser.continuous = false
		return err
	}
	go func() {
		if err := e.watch(); err != nil {
			e.logger("watch", "", err)
		}
	}()

	if e.parser.continuous == false {
		e.parser = parser{[]string{}, false, []string{}, false, 0, []string{}, false, false}
	}
	for {
		text, err := read(nil)
		if err != nil {
			cleanDirs(e.BldDir)
			e.parser.continuous = false
			return err
		}
		e.parser.parseLine(text)
		e.logger("read", text, nil)
		if e.parser.mainClosed {
			e.logger("mainClosed", "true", nil)
			e.parser.continuous = true
			break
		} else {
			e.logger("mainClosed", "false", nil)
		}
	}
	if e.parser.continuous {
		lines := []string{}
		lines = append(lines, convertImport(e.parser.importPkgs)...)
		lines = append(lines, e.parser.body...)
		lines = append(lines, e.parser.main...)
		e.write(lines)
	}
	e.parser.mainClosed = false
	e.parser.main = nil
	return nil
}
