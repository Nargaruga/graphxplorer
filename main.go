package main

import (
	"fmt"
	"log"
	"os"

	"github.com/awalterschulze/gographviz"
)

// Panic in case of error
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if len(os.Args[1:]) < 1 {
		fmt.Println("Usage: ./graphxplorer <FILENAME>")
		os.Exit(1)
	}

	// Read the contents of the file specified by the user
	var path = os.Args[1]
	contents, err := os.ReadFile(path)
	check(err)

	// TODO: check null pointer?
	graph := *parseGraph(contents)
	check(err)

	bfs(graph)
}

// Parse the provided bytes into a DOT graph and return it
func parseGraph(data []byte) *gographviz.Graph {
	graphAst, err := gographviz.Parse(data)
	check(err)
	graph := gographviz.NewGraph()
	err = gographviz.Analyse(graphAst, graph)
	check(err)

	return graph
}
