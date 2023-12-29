package main

import (
	"fmt"
	"log"
	"os"
	"sync"

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

	result_ch := make(chan string)
	done_ch := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		bfs(graph, result_ch, done_ch)
	}()
	go func() {
		defer wg.Done()
		gather_results(result_ch, done_ch)
	}()
	wg.Wait()
	fmt.Println("Finished successfully.")
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

// Gather results from result_ch and print them until a message arrives on done_ch
func gather_results(result_ch chan string, done_ch chan bool) {
	var explored_nodes_list []string

	for {
		select {
		case res := <-result_ch:
			explored_nodes_list = append(explored_nodes_list, res)
		case <-done_ch:
			fmt.Println("Explored", len(explored_nodes_list), "nodes:")
			for _, node := range explored_nodes_list {
				fmt.Println(node)
			}
			return
		}
	}
}
