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
	cleanDirs()
	e := NewEnv(*d)

	e.shell(nil)
}
