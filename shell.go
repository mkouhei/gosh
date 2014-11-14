package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func (e *env) read(fp *os.File, wc, qc chan<- bool, iq chan<- string) {
	go func() {
		o := true
		reader := bufio.NewReader(fp)
		for {
			if o {
				fmt.Print(">>> ")
			} else {
				o = true
			}
			line, _, err := reader.ReadLine()
			if err != nil {
				e.logger("read", "", err)
				cleanDir(e.BldDir)
				qc <- true
				return
			}
			if e.parser.parseLine(string(line), iq) {
				wc <- true
				o = false
			}
		}
	}()
}

func (e *env) write(ic chan<- bool) {

	go func() {
		f, err := os.OpenFile(e.TmpPath, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return
		}
		time.Sleep(time.Microsecond)
		f.Truncate(0)

		for _, l := range e.parser.convertLines() {
			f.WriteString(fmt.Sprintf("%s\n", l))
			e.logger("write", l, nil)
		}
		f.Sync()
		if err := f.Close(); err != nil {
			e.logger("writer", "", err)
			return
		}

		ic <- true
		e.parser.main = nil
	}()
}

func (e *env) goRun(rc chan<- bool) {

	go func() {
		os.Chdir(e.BldDir)
		cmd := "go"
		args := []string{"run", tmpname}
		if msg, err := runCmd(cmd, args...); err != nil {
			e.logger("go run", msg, err)
			e.parser.body = nil
			return
		}
		rc <- true
	}()
}

func (e *env) removeImport(msg, pkg string) {
	if strings.Contains(msg, fmt.Sprintf("package %s: unrecognized import path \"%s\"", pkg, pkg)) {
		removeItem(&e.parser.importPkgs, pkg)
	}
}

func (e *env) goGet(p <-chan string) {
	go func() {
		for {
			pkg := <-p
			cmd := "go"
			args := []string{"get", pkg}
			if msg, err := runCmd(cmd, args...); err != nil {
				e.removeImport(msg, pkg)
				e.logger("go get", msg, err)
			}
			time.Sleep(time.Nanosecond)
		}
	}()
}

func (e *env) goImports(ec chan<- bool) {
	go func() {
		cmd := "goimports"
		args := []string{"-w", e.TmpPath}
		if msg, err := runCmd(cmd, args...); err != nil {
			e.logger("goimports", msg, err)
			e.parser.body = nil
			return
		}
		time.Sleep(time.Nanosecond)
		ec <- true

	}()
}

func (e *env) shell(fp *os.File) {

	if fp == nil {
		fp = os.Stdin
	}

	qc := make(chan bool)
	rc := make(chan bool)
	wc := make(chan bool)
	ic := make(chan bool)
	ec := make(chan bool)
	iq := make(chan string, 10)

	e.read(fp, wc, qc, iq)
	e.goGet(iq)

loop:
	for {
		select {
		case <-wc:
			e.write(ic)
		case <-ic:
			e.goImports(ec)
		case <-ec:
			e.goRun(rc)
		case <-qc:
			cleanDir(e.BldDir)
			fmt.Println("[gosh] terminated")
			break loop
		}
	}

	time.Sleep(time.Nanosecond)
	return
}
