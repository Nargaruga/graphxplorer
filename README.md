# GraphXplorer <!-- omit in toc -->
This is a toy project developed as part of the Principles for Software Composition course @ Unipi, with the purpose of learning the basics of Go's [channels and goroutines](https://go.dev/tour/concurrency/1). 

1. [Features](#features)
2. [Dependencies](#dependencies)
3. [Usage](#usage)
4. [Details About the Parallelized Implementation](#details-about-the-parallelized-implementation)
   1. [Spawn Hierarchy](#spawn-hierarchy)
   2. [Communication Graph](#communication-graph)


## Features
The program takes as input a file describing a graph in the [DOT language](https://graphviz.org/doc/info/lang.html) and a list of nodes. The provided list of nodes is then used as the starting point for the exploration of the graph. When the exploration is complete, a list of all reached nodes is printed out along with their respective distances from (the closest of) the starting nodes.

The exploration is implemented as a BFS whose frontier is initialized to the list of provided starting nodes. There are two implementations of the BFS, ran one after the other for comparison:
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
./graphxplorer [-verbose] [-n_workers n] <path> <starting_node_1> ... <starting_node_n>
```
where
- `path` is the path to a `.gv` file;
- `<starting_node_1> ... <starting_node_n>` is the name of the nodes that make up the starting frontier.

The `verbose` flag can be used to obtain more details about the search.
The `n_workers` flag can be used to specify the number of parallel workers to use in the search.

The `example_inputs` contains a couple of example files that can be used to test the program

## Details About the Parallelized Implementation
The parallel implementation uses no mutexes, instead relying only on channels and goroutines. The search is carried out by `n` worker coroutines, where `n` is provided by the user. Access to the map of node distances and to the frontier is regulated by two dedicated goroutines, which control access to the respective resource through message passing, as depicted in the [communication graph](#communication-graph).

### Spawn Hierarchy
The hierarchy of spawned goroutines is organized as follows:

![spawn_hierarchy](https://i.imgur.com/dK6Q4EZ.png)

### Communication Graph
The communication graph, where nodes are goroutines and edges represent channels, is structured as follows:

![communication_graph](https://i.imgur.com/1iQsCTf.png)
