package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/awalterschulze/gographviz"
)

func main() {
	// Parse flags and arguments
	verbose_ptr := flag.Bool("verbose", false, "print more information about the search")
	n_workers_ptr := flag.Int("n_workers", 1, "number of parallel workers")
	flag.Parse()
	verbose := *verbose_ptr
	n_workers := *n_workers_ptr

	// Check that we got at least the two required arguments
	if len(flag.Args()) < 2 {
		fmt.Println("Usage: ./graphxplorer [-verbose] <path> <starting_node_1> ... <starting_node_n>")
		os.Exit(1)
	}

	// Path to the file to be parsed
	file_path := flag.Args()[0]
	// Names of the nodes in the starting frontier
	starting_node_names := flag.Args()[1:]

	graph, err := deserializeGraph(file_path)
	if err != nil {
		log.Fatal(err)
	}

	if graph == nil || graph.Nodes == nil || len(graph.Nodes.Nodes) == 0 {
		log.Fatal("No graph found.")
	}

	starting_nodes, err := buildInitialFrontier(*graph, starting_node_names)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("--- Sequential ---")
	explore_graph(*graph, BFS, starting_nodes, verbose)

	fmt.Println()

	fmt.Println("--- Parallel ---")
	explore_graph(*graph, ParallelBFS(n_workers), starting_nodes, verbose)
}

// Builds the initial frontier from the requested nodes and returns it
func buildInitialFrontier(graph gographviz.Graph, starting_node_names []string) ([]gographviz.Node, error) {
	var starting_nodes []gographviz.Node

	for _, starting_node_name := range starting_node_names {
		starting_node, ok := graph.Nodes.Lookup[starting_node_name]
		if starting_node == nil || !ok {
			return nil, fmt.Errorf("invalid starting node: %s", starting_node_name)
		}
		starting_nodes = append(starting_nodes, *starting_node)
	}

	return starting_nodes, nil
}

// Parse the contents of the provided file into a DOT graph and return it
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

// Explore the graph with the requested strategy
func explore_graph(graph gographviz.Graph, strategy ExplorationStrategy, starting_nodes []gographviz.Node, verbose bool) {
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
		strategy(graph, starting_nodes, node_data_ch, done_ch)
	}()
	// Start the monitoring function in a new goroutine
	go func() {
		defer wg.Done()
		gather_results(node_data_ch, done_ch, verbose)
	}()
	wg.Wait()
	end := time.Now()
	fmt.Println("Finished successfully in ", end.Sub(start).Microseconds(), "us.")
}

// Gather information about the nodes and print it once the exploration is over
func gather_results(node_data_ch <-chan NodeData, done_ch <-chan bool, verbose bool) {
	node_distances := make(map[string]int)

	for {
		select {
		case node := <-node_data_ch:
			// Update the nodes's distance
			node_distances[node.Name] = node.Dist
		case <-done_ch:
			// Print information about all the received nodes and return
			fmt.Println("Explored", len(node_distances), "nodes.")

			if verbose {
				fmt.Println("Nodes:")

				// List nodes by their distance from the starting frontier
				sorted := listByDistance(node_distances)

				for _, node := range sorted {
					fmt.Println("\t- ", node.Name, "at distance", node.Dist)
				}
			}

			return
		}
	}
}

// List nodes based on their recorded distance, in ascending order
func listByDistance(node_distances map[string]int) []NodeData {
	var sorted []NodeData
	for name, dist := range node_distances {
		sorted = append(sorted, NodeData{Name: name, Dist: dist})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Dist < sorted[j].Dist
	})

	return sorted
}
