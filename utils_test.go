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
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCmd(t *testing.T) {
	cmd := "foo"
	args := []string{}
	if msg, err := runCmd(true, false, cmd, args...); err == nil {
		t.Fatalf("want: <fail>: %s", msg)
	}
	cmd = "true"
	if msg, err := runCmd(true, false, cmd, args...); err != nil {
		t.Fatal(msg)
	}
}

func TestBldDirAndCleanDir(t *testing.T) {
	d := bldDir()
	f, err := os.Stat(d)
	if err != nil {
		t.Fatal(err)
	}
	if !f.IsDir() {
		t.Fatalf("expecting directory: %s", d)
	}
	if err := cleanDir(d); err != nil {
		t.Fatal(err)
	}
	cleanDirs()
	lists, _ := filepath.Glob(fmt.Sprintf("/tmp/%s*", prefix))
	if lists != nil {
		t.Fatal("FAILURE cleanDirs()")
	}

	f2 := bldDir()
	if err := os.Chmod(f2, 0100); err != nil {
		t.Fatal(err)
	}
	if err := cleanDir(f2); err != nil {
		t.Fatal(err)
	}
}

func TestGoVersion(t *testing.T) {
	if !strings.HasPrefix(goVersion(goVer), "go version") {
		t.Fatal("expecting 'go version goX.X.X os/arch'")
	}

	goVer = "go version goX.X.X"
	if !strings.HasPrefix(goVersion(goVer), "go version") {
		t.Fatal("expecting 'go version goX.X.X os/arch'")
	}
}
