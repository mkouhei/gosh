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
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func bldDir() string {
	f, err := ioutil.TempDir("", prefix)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func cleanDir(targetDir string) error {
	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}
	return nil
}

func cleanDirs() {
	lists, _ := filepath.Glob(fmt.Sprintf("/tmp/%s*", prefix))
	for _, l := range lists {
		cleanDir(l)
	}
}

func suppressError(m string) {
	if !strings.HasPrefix(m, "go install: no install location") {
		fmt.Printf("[error] %s", m)
	}
}

func runCmd(printFlag bool, command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		suppressError(stderr.String())
		return stderr.String(), err
	}
	if printFlag {
		fmt.Print(stdout.String())
	}
	return stdout.String(), nil
}

func compare(A, B []string) []string {
	m := make(map[string]int)
	for _, b := range B {
		m[b]++
	}
	var ret []string
	for _, a := range A {
		if m[a] > 0 {
			m[a]--
			continue
		}
		ret = append(ret, a)
	}
	return ret
}

func removeItem(slice *[]string, key string) {
	s := *slice
	for i, item := range s {
		if item == key {
			s = append(s[:i], s[i+1:]...)
		}
	}
	*slice = s
}

func searchString(s string, list []string) bool {
	sort.Strings(list)
	i := sort.SearchStrings(list, s)
	return i < len(list) && list[i] == s
}

func (e *env) logger(facility, msg string, err error) {
	if e.Debug {
		if err == nil {
			log.Printf("[info] %s: %s\n", facility, msg)
		} else {
			log.Printf("[error] %s: %s %v\n", facility, msg, err)
		}
	}
}
