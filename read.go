package main

import (
	"bufio"
	"fmt"
	"os"
)

func reader() ([]byte, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(">>> ")
	text, err := reader.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	return text, nil
}
