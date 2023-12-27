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

	// TODO: check null pointer?
	starting_node := *graph.Nodes.Nodes[0]
	var frontier []gographviz.Node
	frontier = append(frontier, starting_node)
	explored := make(map[string]bool)

	bfs(graph, frontier, explored)
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

// Perform a BFS on the provided graph, printing every explored node
func bfs(graph gographviz.Graph, frontier []gographviz.Node, explored map[string]bool) {
	if len(frontier) == 0 {
		return
	}

	// Pop the head of the frontier
	node := frontier[0]
	frontier = frontier[1:]
	// Skip already explored nodes
	if explored[node.Name] {
		bfs(graph, frontier, explored)
		return
	}

	// Update the frontier with this node's neighbours
	frontier = append(frontier, get_neighbours(graph, node)...)

	// Mark this node as explored to avoid revisiting it
	explored[node.Name] = true
	fmt.Println("Explored:", node.Name)

	bfs(graph, frontier, explored)
}

// Return an array containing all the neighbours of the provided node
func get_neighbours(graph gographviz.Graph, node gographviz.Node) []gographviz.Node {
	var neighbours []gographviz.Node
	// All nodes immediately reachable from this node
	for dst := range graph.Edges.SrcToDsts[node.Name] {
		neighbours = append(neighbours, *graph.Nodes.Lookup[dst])
	}
	// All nodes from which we can immediatey reach this node
	for src := range graph.Edges.DstToSrcs[node.Name] {
		neighbours = append(neighbours, *graph.Nodes.Lookup[src])
	}

	return neighbours
}
