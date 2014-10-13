package main

import (
	"flag"
	"fmt"
)

var version string
var show_version = flag.Bool("version", false, "show_version")

func main() {
	flag.Parse()
	if *show_version {
		fmt.Printf("version: %s\n", version)
		return
	}
}
