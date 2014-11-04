package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/howeyc/fsnotify"
)

func (e *env) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	done := make(chan bool)

	bin := strings.Split(filepath.Clean(e.TmpPath), ".")[0]

	go func() {
		for event := range watcher.Event {
			e.logger("watch", fmt.Sprintf("detect event: %s", event), nil)
			if event.Name == filepath.Clean(e.TmpPath) {
				if event.IsModify() {
					if err := e.goBuild(); err != nil {
						e.logger("go build", "", err)
					}
				}
			} else if event.Name == bin {
				if event.IsCreate() {
					e.logger("go build", bin, nil)
					err := runCmd(bin, []string{}...)
					e.logger(bin, fmt.Sprintf("execute %s", bin), err)
					if err == nil {
						e.resetFile()
					}
				}
			} else {
				e.logger("watch", fmt.Sprintf("unexpected event recieved: %s", event), nil)
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
