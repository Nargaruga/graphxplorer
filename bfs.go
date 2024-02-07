package main

import (
	"fmt"
	"math"
	"sync"

	"github.com/awalterschulze/gographviz"
)

// Holds the name of a node and its distance from the starting nodes
type NodeData struct {
	Name string
	Dist int
}

// Holds the name of the node we are requesting information about and
// the channel on which we will expect the response
type distanceRequest struct {
	Name        string
	Response_ch chan int
}

// Function implementing a graph exploration strategy
type ExplorationStrategy func(gographviz.Graph, []gographviz.Node, chan<- NodeData, chan bool) error

// Perform a BFS on the provided graph, sending data about every new node on the
// 'out_ch' channel and communicating the end of the search through the
// 'done_ch' channel. The `starting_nodes` are used as the starting point for the exploration.
func BFS(graph gographviz.Graph, starting_nodes []gographviz.Node, out_ch chan<- NodeData, done_ch chan bool) error {
	// Maps each node to its distance from the starting nodes
	distances := make(map[string]int)
	// The current node frontier
	var frontier []gographviz.Node
	frontier = append(frontier, starting_nodes...)

	for _, node := range starting_nodes {
		distances[node.Name] = 0
	}

	for len(frontier) != 0 {
		// Pop the head of the frontier
		node := frontier[0]
		frontier = frontier[1:]

		neighbours := getNeighbours(graph, node)
		for _, neighbour := range neighbours {
			_, already_visited := distances[neighbour.Name]
			// If a neighbour hasn't yet been explored, we just found the
			// shortest path to it.
			if !already_visited {
				distances[neighbour.Name] = distances[node.Name] + 1
				frontier = append(frontier, neighbour)
			}
		}

		out_ch <- NodeData{node.Name, distances[node.Name]}
	}

	done_ch <- true
	return nil
}

// Perform a parallelized BFS on the provided graph with n worker goroutines, sending data about every new node on the
// 'out_ch' channel and communicating the end of the search through the 'done_ch' channel
func ParallelBFS(n_workers int) ExplorationStrategy {
	return func(graph gographviz.Graph, starting_nodes []gographviz.Node, out_ch chan<- NodeData, done_ch chan bool) error {
		// Channel to receive the current frontier
		frontier_ch := make(chan []gographviz.Node)
		// Channel to append nodes to the frontier
		frontier_append_ch := make(chan []gographviz.Node)

		// Channel to request the distance of a node
		distance_req_ch := make(chan distanceRequest)
		// Channel to update the distance of a node
		distance_update_ch := make(chan NodeData)

		// Channel to send jobs to the workers
		jobs_ch := make(chan gographviz.Node)

		// Access to the frontier and the distance data is regulated by dedicated goroutines
		go maintainFrontier(starting_nodes, frontier_ch, frontier_append_ch, done_ch)
		go maintainDistances(starting_nodes, distance_req_ch, distance_update_ch, done_ch)

		// Barriers
		var iteration_wg sync.WaitGroup
		var search_wg sync.WaitGroup
		search_wg.Add(n_workers)

		// Spawn `n_workers` explorer goroutines
		for i := 0; i < n_workers; i++ {
			// Each explorer goroutine uses a private channel to receive node distances
			distance_ch := make(chan int)

			go func(id int, distance_ch chan int) {
				processed_nodes := 0

				for {
					// Receive a new node to process
					node, ok := <-jobs_ch
					if !ok {
						break
					}

					// Get the node's neighbours
					neighbours := getNeighbours(graph, node)
					var neighbours_to_update []gographviz.Node

					// Request the node's distance
					distance_req_ch <- distanceRequest{Name: node.Name, Response_ch: distance_ch}
					current_node_dist := <-distance_ch

					// Process each neighbour
					for _, neighbour := range neighbours {
						distance_req_ch <- distanceRequest{Name: neighbour.Name, Response_ch: distance_ch}
						old_dist := <-distance_ch
						new_dist := current_node_dist + 1

						// We only update the distance if it's shorter than the currently recorded one
						if new_dist < old_dist {
							distance_update_ch <- NodeData{neighbour.Name, new_dist}
							neighbours_to_update = append(neighbours_to_update, neighbour)
						}
					}

					// Update the frontier
					frontier_append_ch <- neighbours_to_update
					// Send info about the node we just explored
					out_ch <- NodeData{node.Name, current_node_dist}

					processed_nodes++
					iteration_wg.Done()
				}

				fmt.Println("Worker", id, "processed", processed_nodes, "nodes.")
				search_wg.Done()
			}(i, distance_ch)
		}

		for {
			// Extract the current frontier
			current_frontier := <-frontier_ch

			// No more nodes in the frontier: we are done
			if len(current_frontier) == 0 {
				break
			}

			// Distribute jobs
			for _, node := range current_frontier {
				iteration_wg.Add(1)
				jobs_ch <- node
			}

			// We need to wait for all goroutines to append their newly found nodes
			// to the frontier in order to avoid a premature exit
			iteration_wg.Wait()
		}

		close(jobs_ch)
		search_wg.Wait()

		close(done_ch)
		return nil
	}
}

// Return an array containing all the neighbours of the provided node
func getNeighbours(graph gographviz.Graph, node gographviz.Node) []gographviz.Node {
	var neighbours []gographviz.Node

	// All nodes reachable from this node in one step
	for dst := range graph.Edges.SrcToDsts[node.Name] {
		neighbours = append(neighbours, *graph.Nodes.Lookup[dst])
	}

	if !graph.Directed {
		// All nodes from which this node can be reached in one step
		for src := range graph.Edges.DstToSrcs[node.Name] {
			neighbours = append(neighbours, *graph.Nodes.Lookup[src])
		}
	}

	return neighbours
}

// Manage access to the frontier, appending received nodes and sending the current state of the frontier when requested
func maintainFrontier(starting_nodes []gographviz.Node, frontier_ch chan<- []gographviz.Node, frontier_append_ch <-chan []gographviz.Node, done_ch <-chan bool) {
	// The current node frontier
	var frontier []gographviz.Node
	frontier = append(frontier, starting_nodes...)

	for {
		select {
		case frontier_ch <- frontier:
			// Send the current frontier
			frontier = nil
		case nodes := <-frontier_append_ch:
			// Append the received nodes to the frontier
			frontier = append(frontier, nodes...)
		case <-done_ch:
			return
		}
	}
}

// Manage access to the recorded node distances, handling queries and updates
func maintainDistances(starting_nodes []gographviz.Node, distance_req_ch <-chan distanceRequest, distance_update_ch <-chan NodeData, done_ch <-chan bool) {
	// Maps each node to its distance from the starting nodes
	distances := make(map[string]int)

	for _, node := range starting_nodes {
		distances[node.Name] = 0
	}

	for {
		select {
		case req := <-distance_req_ch:
			// Send the requested node distance
			dist, ok := distances[req.Name]
			if !ok {
				// This is the first time seeing the node, so we set its distance to the maximum
				dist = math.MaxInt32
			}
			req.Response_ch <- dist
		case node_data := <-distance_update_ch:
			// Update the distance with the provided data
			distances[node_data.Name] = node_data.Dist
		case <-done_ch:
			return
		}
	}
}
