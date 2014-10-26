package main

import (
	"os"
	"time"
)

func writeFile(codePath string, content string) error {
	f, err := os.OpenFile(codePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	time.Sleep(time.Microsecond)

	f.WriteString(content)
	f.Sync()
	f.Close()

	return nil
}
