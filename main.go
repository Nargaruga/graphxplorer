package main

import (
	"fmt"
	"gonum.org/v1/gonum/graph/formats/dot"
	"os"
)

// Panics if provided with an error
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	if len(os.Args[1:]) < 1 {
		fmt.Println("Usage: ./graphxplorer <FILENAME>")
		os.Exit(1)
	}

	var path = os.Args[1]
	dot_file, err := dot.ParseFile(path)
	check(err)
	fmt.Println(dot_file.String())
}
