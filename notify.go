package main

import (
	"fmt"

	"github.com/howeyc/fsnotify"
)

func watch(targetDir string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("new watcher error:", err)
		return err
	}
	done := make(chan bool)

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				fmt.Println("event: ", ev)
			case err := <-watcher.Error:
				fmt.Println("error: ", err)
				break
			}
		}
	}()

	err = watcher.Watch(targetDir)
	if err != nil {
		fmt.Println("watch error:", err)
		return err
	}

	<-done

	fmt.Print("finished")

	watcher.Close()
	return nil
}
