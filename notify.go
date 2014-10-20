package main

import (
	"fmt"

	"path/filepath"
	"sync/atomic"

	"os"

	"github.com/howeyc/fsnotify"
)

const (
	tmpname = "gosh_tmp.go"
)

type cnt struct {
	val int32
}

func (c *cnt) incremant() {
	atomic.AddInt32(&c.val, 1)
}

func goBuild(bldDir, codePath string) error {
	os.Chdir(bldDir)
	cmd := "go"
	args := []string{"build", codePath}
	if err := runCmd(cmd, args...); err != nil {
		return err
	}
	return nil
}

func watch(targetDir string) error {
	tmpFile := fmt.Sprintf("%s/%s", targetDir, tmpname)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("new watcher error:", err)
		return err
	}
	var modifyRecieved cnt
	done := make(chan bool)

	go func() {
		for event := range watcher.Event {
			fmt.Println(event)
			if event.Name == filepath.Clean(tmpFile) {
				fmt.Printf("event recieved: %s", event)
				if event.IsModify() {
					modifyRecieved.incremant()
					goBuild(targetDir, tmpFile)
				}
			} else {
				fmt.Printf("unexpected event recieved: %s", event)
			}
		}
		done <- true
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
