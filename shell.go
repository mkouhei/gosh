package main

/*
   Copyright (C) 2014 Kouhei Maeda <mkouhei@palmtb.net>

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

func (e *env) goRun() {

	go func() {
		os.Chdir(e.BldDir)
		cmd := "go"
		args := []string{"run", tmpname}
		if msg, err := runCmd(true, cmd, args...); err != nil {
			e.logger("go run", msg, err)
			e.parser.body = nil
			return
		}
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
			if msg, err := runCmd(true, cmd, args...); err != nil {
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
		if msg, err := runCmd(true, cmd, args...); err != nil {
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
			e.goRun()
		case <-qc:
			cleanDir(e.BldDir)
			fmt.Println("[gosh] terminated")
			break loop
		}
	}

	time.Sleep(time.Nanosecond)
	return
}
