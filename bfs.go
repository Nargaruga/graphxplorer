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

// TODO: desc
func parallel_bfs(graph gographviz.Graph, out_ch chan string, done_ch chan bool) {
	// TODO: check null pointer?
	starting_node := *graph.Nodes.Nodes[0]
	explored := make(map[string]bool)
	var frontier []gographviz.Node
	frontier = append(frontier, starting_node)

	req_frontier_ch := make(chan bool)
	frontier_ch := make(chan []gographviz.Node)
	append_ch := make(chan []gographviz.Node)
	cleanup_frontier_ch := make(chan bool)

	go maintain_frontier(frontier, req_frontier_ch, frontier_ch, append_ch, cleanup_frontier_ch)

	var wg sync.WaitGroup

	for {
		// Request the current frontier
		req_frontier_ch <- true
		current_frontier := <-frontier_ch

		if len(current_frontier) == 0 {
			break
		}

		for _, node := range current_frontier {
			if explored[node.Name] {
				continue
			}
			explored[node.Name] = true

			// Each node is processed in a separate goroutine
			wg.Add(1)
			go func(to_process gographviz.Node) {
				defer wg.Done()
				append_ch <- get_neighbours(graph, to_process)
				out_ch <- to_process.Name
			}(node)
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

// Manage access to the frontier, appending received nodes and sending the current state of the frontier when requested
func maintain_frontier(frontier []gographviz.Node, req_frontier_ch chan bool, frontier_ch chan []gographviz.Node, append_ch chan []gographviz.Node, done_ch chan bool) {
	for {
		select {
		case <-req_frontier_ch:
			frontier_ch <- frontier
			frontier = nil
		case nodes := <-append_ch:
			frontier = append(frontier, nodes...)
		case <-done_ch:
			return
		}
	}
}
