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
	"fmt"
	"os"
)

type env struct {
	bldDir  string
	tmpName string
	tmpPath string
	goPath  string
	debug   bool

	parserSrc parserSrc
	readFlag  int32
}

func newEnv(debug bool) env {
	// New shell environment
	e := env{}
	e.bldDir = bldDir()
	e.tmpPath = fmt.Sprintf("%s/%s", e.bldDir, tmpname)
	e.goPath = e.bldDir
	e.debug = debug
	e.parserSrc = parserSrc{}
	e.readFlag = 0

	setGOPATH(e.bldDir)
	e.logger("GOPATH", os.Getenv("GOPATH"), nil)
	return e
}

func setGOPATH(p string) {
	// set GOPATH
	os.Setenv("GOPATH", p)
}
