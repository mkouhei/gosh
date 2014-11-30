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
	"testing"
)

func TestNewEnv(t *testing.T) {
	e := newEnv(false)
	_, err := os.Stat(e.BldDir)
	if err != nil {
		t.Fatal(err)
	}
	if e.TmpPath != fmt.Sprintf("%s/%s", e.BldDir, tmpname) {
		t.Fatal("error initialize")
	}
	if e.GoPath != e.BldDir {
		t.Fatal("error initialize")
	}
	if e.Debug {
		t.Fatal("error initialize")
	}
	cleanDir(e.BldDir)
}
