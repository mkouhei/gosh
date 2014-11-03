package main

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
}

func NewEnv(debug bool) env {
	e := env{}
	e.BldDir = bldDir()
	e.TmpPath = fmt.Sprintf("%s/%s", e.BldDir, tmpname)
	e.GoPath = e.BldDir
	e.Debug = debug
	setGOPATH(e.BldDir)
	e.logger("GOPATH", os.Getenv("GOPATH"), nil)
	return e
}

func setGOPATH(p string) {
	os.Setenv("GOPATH", p)
}
