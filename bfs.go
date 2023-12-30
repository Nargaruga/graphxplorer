package main

import (
	"sync"

	"github.com/awalterschulze/gographviz"
)

// Perform a BFS on the provided graph, sending every new node on the 'out_ch' channel
func bfs(graph gographviz.Graph, out_ch chan string, done_ch chan bool) {
	// TODO: check null pointer?
	starting_node := *graph.Nodes.Nodes[0]
	explored := make(map[string]bool)
	var frontier []gographviz.Node
	frontier = append(frontier, starting_node)

	for len(frontier) != 0 {
		// Pop the head of the frontier
		node := frontier[0]
		frontier = frontier[1:]

		// Skip node if already explored
		if explored[node.Name] {
			continue
		}

		// Mark this node as explored to avoid revisiting it
		explored[node.Name] = true

		// Update the frontier with this node's neighbours
		neighbours := get_neighbours(graph, node)
		frontier = append(frontier, neighbours...)

		out_ch <- node.Name
	}

	done_ch <- true
}

// Perform a parallelized BFS on the provided graph, sending every new node on the 'out_ch' channel
func parallel_bfs(graph gographviz.Graph, out_ch chan string, done_ch chan bool) {
	// TODO: check null pointer?
	starting_node := *graph.Nodes.Nodes[0]
	explored := make(map[string]bool)
	var frontier []gographviz.Node
	frontier = append(frontier, starting_node)

	var frontier_mx sync.Mutex
	var explored_mx sync.Mutex

	var wg sync.WaitGroup
	for len(frontier) != 0 {
		frontier_size := len(frontier)
		wg.Add(frontier_size)
		for i := 0; i < frontier_size; i++ {
			go func() {
				defer wg.Done()

				// Pop the head of the frontier
				frontier_mx.Lock()
				node := frontier[0]
				frontier = frontier[1:]
				frontier_mx.Unlock()

				// Skip node if already explored
				explored_mx.Lock()
				already_visited := explored[node.Name]
				if already_visited {
					explored_mx.Unlock()
					return
				}

				// Mark this node as explored to avoid revisiting it
				explored[node.Name] = true
				explored_mx.Unlock()

				// Update the frontier with this node's neighbours
				neighbours := get_neighbours(graph, node)
				frontier_mx.Lock()
				frontier = append(frontier, neighbours...)
				frontier_mx.Unlock()

				out_ch <- node.Name
			}()
		}
		wg.Wait()
	}

	done_ch <- true
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
