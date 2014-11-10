package main

import (
	"os"
)

func ExampleMain() {
	os.Args = append(os.Args, "-version")
	main()
	// Output:
	// version:
}
