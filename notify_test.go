package main

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestWatch(t *testing.T) {

	blddir := bldDir()
	tmpFile := fmt.Sprintf("%s/%s", blddir, "testcode")

	go func() {
		if err := watch(blddir); err != nil {
			t.Fatal(err)
		}
	}()

	err := ioutil.WriteFile(tmpFile, []byte{}, 0644)
	if err != nil {
		t.Fatal(err)
	}

}
