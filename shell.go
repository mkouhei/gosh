package main

import (
	"fmt"
)

func shell() {
	for {
		text, err := reader()
		if err != nil {
			break
		}
		fmt.Print(string(text))
		if err := writeFile(tmpname, string(text)); err != nil {
			fmt.Printf("[error] %v", err)
			break
		}
	}
}
