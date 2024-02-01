# GraphXplorer
This is a toy project developed with the sole purpose of learning the basics of Go's [channels and goroutines](https://go.dev/tour/concurrency/1). 

## Features
The program takes as input a file describing a graph in the [DOT language](https://graphviz.org/doc/info/lang.html) as well as the name of a node. It then performs a BFS on the graph starting from the provided node and prints out a list of all reachable nodes along with their distance from the starting point.

Two breadth first searches are carried out:
- a sequential search, implemented in the typical, iterative fashion;
- a parallelized BFS which makes use of Go's channels and goroutines.

## Dependencies
- [gographviz](https://github.com/awalterschulze/gographviz)

## Usage
Navigate to the project's root and build it with
```
go build
```

The program can be started as 
```
./graphxplorer [-verbose] <path> <starting_node_1> ... <starting_node_n>
```
where
- `path` is the path to a `.gv` file;
- `<starting_node_1> ... <starting_node_n>` is the name of the nodes that make up the starting frontier.

The `verbose` flag can be used to obtain more details about the search.

The `example_inputs` contains a couple of example files that can be used to test the program

## Structure

### Spawn Hierarchy
The hierarchy of spawned goroutines is organized as follows:

![spawn_hierarchy](https://i.imgur.com/zZ0Xnaq.png)

### Communication Graph
The communication graph, where nodes are goroutines and edges represent channels, is structured as follows:

![communication_graph](https://i.imgur.com/n3lXkrx.png)