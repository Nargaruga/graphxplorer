package main

import (
	"errors"
	"math"
	"sync"

	"github.com/awalterschulze/gographviz"
)

// Holds the name of a node and its distance from the starting node
type NodeData struct {
	Name string
	Dist int
}

// Perform a BFS on the provided graph, sending data about every new node on the
// 'out_ch' channel and communicating the end of the search through the
// 'done_ch' channel
func BFS(graph gographviz.Graph, out_ch chan NodeData, done_ch chan bool) error {
	if graph.Nodes == nil || len(graph.Nodes.Nodes) == 0 || graph.Nodes.Nodes[0] == nil {
		return errors.New("Missing a starting node for the search.")
	}

	starting_node := *graph.Nodes.Nodes[0]

	// Keeps track of already-explored nodes
	explored := make(map[string]bool)
	// Maps each node to its distance from the starting node
	distances := make(map[string]int)
	// The current node frontier
	var frontier []gographviz.Node
	frontier = append(frontier, starting_node)

	distances[starting_node.Name] = 0

	for len(frontier) != 0 {
		// Pop the head of the frontier
		node := frontier[0]
		frontier = frontier[1:]

		if explored[node.Name] {
			continue
		}

		// Only update the distance if it is smaller than the currently recorded one
		neighbours := getNeighbours(graph, node)
		var neighbours_to_update []gographviz.Node
		for _, neighbour := range neighbours {
			old_dist, ok := distances[neighbour.Name]
			new_dist := distances[node.Name] + 1

			// We only update the distance if it's shorter than the one currently recorder
			if !ok || new_dist < old_dist {
				distances[neighbour.Name] = new_dist
				neighbours_to_update = append(neighbours_to_update, neighbour)
			}
		}
		frontier = append(frontier, neighbours_to_update...)

		explored[node.Name] = true
		out_ch <- NodeData{node.Name, distances[node.Name]}
	}

	done_ch <- true
	return nil
}

// Perform a BFS on the provided graph, sending data about every new node on the
// 'out_ch' channel and communicating the end of the search through the
// 'done_ch' channel
func ParallelBFS(graph gographviz.Graph, out_ch chan NodeData, done_ch chan bool) error {
	if graph.Nodes == nil || len(graph.Nodes.Nodes) == 0 || graph.Nodes.Nodes[0] == nil {
		return errors.New("Missing a starting node for the search.")
	}

	starting_node := *graph.Nodes.Nodes[0]

	// Keeps track of already-explored nodes
	explored := make(map[string]bool)

	// Channel to request the current frontier
	req_frontier_ch := make(chan bool)
	// Channel to receive the current frontier
	frontier_ch := make(chan []gographviz.Node)
	// Channel to append nodes to the frontier
	append_ch := make(chan []gographviz.Node)

	// Channel to request the distance of a node
	req_distance_ch := make(chan string)
	// Channel to receive the distance of a node
	distance_ch := make(chan int)
	// Channel to update the distance of a node
	update_distance_ch := make(chan NodeData)

	// Access to the frontier and the distance data are handled by dedicated goroutines
	go maintainFrontier(starting_node, req_frontier_ch, frontier_ch, append_ch, done_ch)
	go maintainDistances(starting_node, req_distance_ch, distance_ch, update_distance_ch, done_ch)

	var wg sync.WaitGroup

	for {
		// Request the current frontier
		req_frontier_ch <- true
		current_frontier := <-frontier_ch

		// No more nodes in the frontier: we are done
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
				neighbours := getNeighbours(graph, to_process)
				var neighbours_to_update []gographviz.Node

				req_distance_ch <- to_process.Name
				current_node_dist := <-distance_ch

				for _, neighbour := range neighbours {
					req_distance_ch <- neighbour.Name
					old_dist := <-distance_ch
					new_dist := current_node_dist + 1

					// We only update the distance if it's shorter than the one currently recorder
					if new_dist < old_dist {
						update_distance_ch <- NodeData{neighbour.Name, new_dist}
						neighbours_to_update = append(neighbours_to_update, neighbour)
					}
				}
				append_ch <- neighbours_to_update

				out_ch <- NodeData{to_process.Name, current_node_dist}
			}(node)
		}

		// We need to wait for all goroutines to append their newly found nodes
		// to the frontier in order to avoid a premature exit
		wg.Wait()
	}

	close(done_ch)
	return nil
}

// Return an array containing all the neighbours of the provided node
func getNeighbours(graph gographviz.Graph, node gographviz.Node) []gographviz.Node {
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
func maintainFrontier(starting_node gographviz.Node, req_frontier_ch chan bool, frontier_ch chan []gographviz.Node, append_ch chan []gographviz.Node, done_ch chan bool) {
	// The current node frontier
	var frontier []gographviz.Node
	frontier = append(frontier, starting_node)

	for {
		select {
		case <-req_frontier_ch:
			// Send the frontier, as requested
			frontier_ch <- frontier
			frontier = nil
		case nodes := <-append_ch:
			// Append the received nodes to the frontier
			frontier = append(frontier, nodes...)
		case <-done_ch:
			return
		}
	}
}

func maintainDistances(starting_node gographviz.Node, req_distance_ch chan string, distance_ch chan int, update_distance_ch chan NodeData, done_ch chan bool) {
	// Maps each node to its distance from the starting node
	distances := make(map[string]int)
	distances[starting_node.Name] = 0

	for {
		select {
		case node_name := <-req_distance_ch:
			// Send the requested node's distance
			dist, ok := distances[node_name]
			if !ok {
				// This is the first time seeing the node, so we set its distance to the maximum
				dist = math.MaxInt32
			}
			distance_ch <- dist
		case node_data := <-update_distance_ch:
			// Update the distance with the provided data
			distances[node_data.Name] = node_data.Dist
		case <-done_ch:
			return
		}
	}
}
