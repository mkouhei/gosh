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
	"flag"
	"fmt"
)

var goVer string
var ver string
var showVersion = flag.Bool("version", false, "showVersion")

var license = `Gosh %s
Copyright (C) 2014,2015 Kouhei Maeda
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>.
This is free software, and you are welcome to redistribute it.
There is NO WARRANTY, to the extent permitted by law.
`

func run(d bool, s string) {
	e := newEnv(d, s)
	e.chkDeps()
	fmt.Println(e.goVersion(goVer))
	fmt.Printf(license, ver)
	e.shell(nil)
}

func main() {
	d := flag.Bool("d", false, "debug mode")
	c := flag.Bool("c", false, "cleanup all Gosh's temporary files")
	s := flag.String("s", "", `specify sudo password for running "sudo -E go run"`)
	flag.Parse()
	if *c {
		cleanDirs()
	}
	if ver == "" {
		ver = version
	}
	if *showVersion {
		fmt.Printf("version: %s\n", ver)
		return
	}
	run(*d, *s)
}
