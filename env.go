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
}

func NewEnv() env {
	e := env{}
	e.BldDir = bldDir()
	e.TmpPath = fmt.Sprintf("%s/%s", e.BldDir, tmpname)
	e.GoPath = e.BldDir
	return e
}

func setGOPATH(p string) {
	os.Setenv("GOPATH", p)
}
