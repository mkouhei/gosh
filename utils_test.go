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
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCmd(t *testing.T) {
	cmd := "foo"
	args := []string{}
	if msg, err := runCmd(true, cmd, args...); err == nil {
		t.Fatal("want: <fail>: %s", msg)
	}
	cmd = "true"
	if msg, err := runCmd(true, cmd, args...); err != nil {
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
}

func TestSearchString(t *testing.T) {
	list := []string{"foo", "bar", "baz"}
	if !searchString("foo", list) {
		t.Fatal("expecting true")
	}
	if searchString("hoge", list) {
		t.Fatal("expecting false")
	}
}

func TestRemoveItem(t *testing.T) {

	slc := []string{"foo"}
	slc2 := []string{}
	removeItem(&slc, "foo")
	if len(compare(slc, slc2)) != 0 {
		t.Fatal("fail remove item.")
	}

	if !searchString("foo", slc) {
		slc = append(slc, "foo")
	}
	slc2 = []string{}
	removeItem(&slc, "foo")
	if len(compare(slc, slc2)) != 0 {
		t.Fatal("fail remove item.")
	}

	slc = []string{"foo", "bar", "baz"}
	slc2 = []string{"foo", "baz"}
	removeItem(&slc, "bar")
	if len(compare(slc, slc2)) != 0 {
		t.Fatal("fail remove item.")
	}

	slc = []string{"foo", "bar", "baz", "qux", "quux"}
	slc2 = []string{"bar", "baz", "qux", "quux"}
	removeItem(&slc, "foo")
	if len(compare(slc, slc2)) != 0 {
		t.Fatal("fail remove item.")
	}

	slc = []string{"foo", "bar", "baz", "qux", "quux"}
	slc2 = []string{"foo", "bar", "baz", "qux"}
	removeItem(&slc, "quux")
	if len(compare(slc, slc2)) != 0 {
		t.Fatal("fail remove item.")
	}

}

func TestGoVersion(t *testing.T) {
	if !strings.HasPrefix(goVersion(), "go version go") {
		t.Fatal("expecting 'go version goX.X.X os/arch'")
	}
}
