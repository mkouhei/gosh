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
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewEnv(t *testing.T) {
	e := newEnv(false, "")
	e.chkDeps()

	_, err := os.Stat(e.bldDir)
	if err != nil {
		t.Fatal(err)
	}
	if e.tmpPath != filepath.Join(e.bldDir, tmpname) {
		t.Fatal("error initialize")
	}
	if e.goPath != e.bldDir {
		t.Fatal("error initialize")
	}
	if e.debug {
		t.Fatal("error initialize")
	}
	cleanDir(e.bldDir)
}

func CheckInst() {
	p := os.Getenv("PATH")
	os.Setenv("PATH", "")
	go checkInst()
	time.Sleep(time.Second)
	d := bldDir()
	t := filepath.Join(d, "goimports")
	f, _ := os.OpenFile(t, os.O_CREATE, 0600)
	f.WriteString("")
	f.Sync()
	f.Close()
	os.Setenv("PATH", p)
	// Output:
	// Please wait for installing "goimports".
}
