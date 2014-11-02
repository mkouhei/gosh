package main

import (
	"log"
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
					goBuild(e.BldDir, e.TmpPath)
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
