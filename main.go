package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/awalterschulze/gographviz"
)

type explorationStrategy func(gographviz.Graph, chan NodeData, chan bool) error

func main() {
	if len(os.Args[1:]) < 1 {
		fmt.Println("Usage: ./graphxplorer <FILENAME>")
		os.Exit(1)
	}

	graph, err := deserializeGraph(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if graph == nil {
		log.Fatal("No graph found.")
	}

	fmt.Println("--- Sequential ---")
	explore_graph(*graph, BFS)

	fmt.Println()

	fmt.Println("--- Parallel ---")
	explore_graph(*graph, ParallelBFS)
}

// Parse the provided bytes into a DOT graph and return it
func deserializeGraph(path string) (*gographviz.Graph, error) {
	// Read file contents
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse the file contents into an AST
	graphAst, err := gographviz.Parse(contents)
	if err != nil {
		return nil, err
	}

	// Create a graph from the AST
	graph := gographviz.NewGraph()
	err = gographviz.Analyse(graphAst, graph)
	if err != nil {
		return nil, err
	}

	return graph, nil
}

// Explore the graph with the provided strategy
func explore_graph(graph gographviz.Graph, strategy explorationStrategy) {
	// Channel for communicating data about the explored nodes
	node_data_ch := make(chan NodeData)
	// Channel for communicating the end of the exploration
	done_ch := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(2)
	start := time.Now()
	// Start the exploration in a new goroutine
	go func() {
		defer wg.Done()
		strategy(graph, node_data_ch, done_ch)
	}()
	// Start the monitoring function in a new goroutine
	go func() {
		defer wg.Done()
		gather_results(node_data_ch, done_ch)
	}()
	wg.Wait()
	end := time.Now()
	fmt.Println("Finished successfully in ", end.Sub(start).Microseconds(), "us.")
}

// Gather information about the nodes and print it once the exploration is over
func gather_results(node_data_ch chan NodeData, done_ch chan bool) {
	node_distances := make(map[string]int)

	for {
		select {
		case node := <-node_data_ch:
			// Update the distance for the newly received node
			node_distances[node.Name] = node.Dist
		case <-done_ch:
			// Print information about all the received nodes and return
			fmt.Println("Explored", len(node_distances), "nodes.")
			fmt.Println("Nodes:")

			// Sort nodes by their distance from the starting node
			var sorted []NodeData
			for name, dist := range node_distances {
				sorted = append(sorted, NodeData{name, dist})
			}
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Dist < sorted[j].Dist
			})

			for _, node := range sorted {
				fmt.Println(node.Name, "at distance", node.Dist)
			}
			return
		}
	}
}
