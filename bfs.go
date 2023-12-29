package main

import (
	"fmt"

	"github.com/awalterschulze/gographviz"
)

// Perform a BFS on the provided graph, printing every explored node
func bfs(graph gographviz.Graph) {
	// TODO: check null pointer?
	starting_node := *graph.Nodes.Nodes[0]
	explored := make(map[string]bool)
	var frontier []gographviz.Node
	frontier = append(frontier, starting_node)

	for len(frontier) != 0 {
		// Pop the head of the frontier
		node := frontier[0]
		frontier = frontier[1:]

		// Skip already explored nodes
		if explored[node.Name] {
			continue
		}

		// Update the frontier with this node's neighbours
		frontier = append(frontier, get_neighbours(graph, node)...)

		// Mark this node as explored to avoid revisiting it
		explored[node.Name] = true
		fmt.Println("Explored:", node.Name)
	}
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
