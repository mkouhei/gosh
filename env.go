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
	"fmt"
	"os"
)

type env struct {
	BldDir  string
	TmpName string
	TmpPath string
	GoPath  string
	Debug   bool

	parser parser
}

func NewEnv(debug bool) env {
	e := env{}
	e.BldDir = bldDir()
	e.TmpPath = fmt.Sprintf("%s/%s", e.BldDir, tmpname)
	e.GoPath = e.BldDir
	e.Debug = debug
	e.parser = parser{}

	setGOPATH(e.BldDir)
	e.logger("GOPATH", os.Getenv("GOPATH"), nil)
	return e
}

func setGOPATH(p string) {
	os.Setenv("GOPATH", p)
}
