package main

import (
	"flag"
	"fmt"
)

var version string
var show_version = flag.Bool("version", false, "show_version")

func main() {
	d := flag.Bool("d", false, "debug mode")
	flag.Parse()
	if *show_version {
		fmt.Printf("version: %s\n", version)
		return
	}
	e := NewEnv(*d)
	for {
		if e.shell() == false {
			break
		}
	}
}
