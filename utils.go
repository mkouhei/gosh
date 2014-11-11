package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

func (e *env) goGet(p string) {
	go func() {
		for {
			cmd := "go"
			args := []string{"get", p}
			if err := runCmd(cmd, args...); err != nil {
				e.logger("go get", "", err)
				return
			}
			time.Sleep(time.Nanosecond)
		}
	}()
}

func (e *env) goImports(ec chan<- bool) {
	go func() {
		cmd := "goimports"
		args := []string{"-w", e.TmpPath}
		if err := runCmd(cmd, args...); err != nil {
			e.logger("goimports", "", err)
			e.parser.body = nil
			return
		}
		time.Sleep(time.Nanosecond)
		ec <- true

	}()
}

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

func suppressError(m string) {
	if !strings.HasPrefix(m, "go install: no install location") {
		fmt.Printf("[error] %s", m)
	}
}

func runCmd(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		suppressError(stderr.String())
		return err
	}
	fmt.Print(stdout.String())
	return nil
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
			log.Fatalf("[error] %s: %s %v\n", facility, msg, err)
		}
	}
}
