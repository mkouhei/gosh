package main

import (
	"bufio"
	"fmt"
	"os"
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
		if err := runCmd(cmd, args...); err != nil {
			e.logger("go run", "", err)
			e.parser.body = nil
			return
		}
		rc <- true
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

	iq <- ""
	e.read(fp, wc, qc, iq)
	goGet(<-iq)

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
