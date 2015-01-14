package main

/*
   Copyright (C) 2014,2015 Kouhei Maeda <mkouhei@palmtb.net>

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
)

func (e *env) read(fp *os.File, wc, qc chan<- bool, iq chan<- importSpec) {
	// read from shell prompt
	go func() {
		reader := bufio.NewReader(fp)
		for {
			if e.readFlag == 0 {
				fmt.Print(">>> ")
			} else {
				e.readFlag--
			}
			line, _, err := reader.ReadLine()
			if err != nil {
				e.logger("read", "", err)
				cleanDir(e.bldDir)
				qc <- true
				return
			}

			// append token.SEMICOLON
			line = append(line, 59)

			if e.parserSrc.parseLine(line, iq) {
				wc <- true
				e.readFlag = 3
			}
		}
	}()
}

func (e *env) write(ic chan<- bool) {
	// write tmporary source code file
	go func() {
		f, err := os.OpenFile(e.tmpPath, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return
		}
		f.Truncate(0)

		for _, l := range e.parserSrc.mergeLines() {
			f.WriteString(fmt.Sprintf("%s\n", l))
			e.logger("write", l, nil)
		}
		f.Sync()
		if err := f.Close(); err != nil {
			e.logger("writer", "", err)
			return
		}

		ic <- true
		e.parserSrc.main = nil
		removePrintStmt(&e.parserSrc.mainHist)
	}()
}

func (e *env) goRun() {
	// execute `go run'
	os.Chdir(e.bldDir)
	cmd := "go"
	args := []string{"run", tmpname}
	omitFlag := false
	if len(e.parserSrc.mainHist) > 0 {
		omitFlag = true
	}
	if msg, err := runCmd(true, omitFlag, cmd, args...); err != nil {
		e.logger("go run", msg, err)
		e.parserSrc.body = nil
		return
	}
}

func (e *env) removeImport(msg string, pkg importSpec) {
	// remove package from env.parser.imPkg
	var key string
	if pkg.pkgName == "" {
		key = pkg.imPath
	} else {
		key = pkg.pkgName
	}
	if strings.Contains(msg, fmt.Sprintf(`package %s: unrecognized import path "%s"`, key, key)) {
		removeImportPackage(&e.parserSrc.imPkgs, importSpec{pkg.imPath, pkg.pkgName})
	}
}

func (e *env) goGet(p <-chan importSpec) {
	// execute `go get'
	go func() {
		for {
			pkg := <-p
			cmd := "go"
			args := []string{"get", pkg.imPath}
			if msg, err := runCmd(true, false, cmd, args...); err != nil {
				e.removeImport(msg, pkg)
				e.logger("go get", msg, err)
			}
		}
	}()
}

func (e *env) goImports(ec chan<- bool) {
	// execute `goimports'
	go func() {
		cmd := "goimports"
		args := []string{"-w", e.tmpPath}
		if msg, err := runCmd(true, false, cmd, args...); err != nil {
			e.logger("goimports", msg, err)
			e.parserSrc.body = nil
		}
		ec <- true

	}()
}

func (e *env) shell(fp *os.File) {
	// main shell loop

	if fp == nil {
		fp = os.Stdin
	}

	// quit channel
	qc := make(chan bool)
	// write channel
	wc := make(chan bool)
	// import channel
	ic := make(chan bool)
	// execute channel
	ec := make(chan bool)
	// package queue for go get
	iq := make(chan importSpec, 10)

	e.goGet(iq)

loop:
	for {
		e.read(fp, wc, qc, iq)

		select {
		case <-wc:
			e.write(ic)
		case <-ic:
			e.goImports(ec)
		case <-ec:
			e.goRun()
		case <-qc:
			cleanDir(e.bldDir)
			fmt.Println("[gosh] terminated")
			break loop
		}
	}

	return
}
