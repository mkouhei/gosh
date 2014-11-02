package main

import (
	"log"

	"strings"

	"path/filepath"

	"github.com/howeyc/fsnotify"
)

func (e *env) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	done := make(chan bool)

	go func() {
		for event := range watcher.Event {
			log.Println(event)
			if event.Name == filepath.Clean(e.TmpPath) {
				if event.IsModify() {
					if err := e.goBuild(); err != nil {
						log.Println("[error] go build: %v\n", err)
					}
				}
			} else if event.Name == strings.Split(filepath.Clean(e.TmpPath), ".")[0] {
				if event.IsCreate() {
					log.Println("[success] go build: %v\n", strings.Split(tmpname, ".")[0])
				}
			} else {
				log.Printf("unexpected event recieved: %s", event)
				break
			}
		}
		done <- true
	}()

	err = watcher.Watch(e.BldDir)
	if err != nil {
		return err
	}
	<-done

	watcher.Close()
	return nil
}
