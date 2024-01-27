# GraphXplorer
This is a toy project developed with the sole purpose of learning the basics of Go's [channels and goroutines](https://go.dev/tour/concurrency/1). 

## Features
The program takes as input a file describing a graph in the [DOT language](https://graphviz.org/doc/info/lang.html) as well as the name of a node. It then performs a BFS on the graph starting from the provided node and prints out a list of all reacheable nodes along with their distance from the starting point.

Two breadth first searches are carried out:
- a sequential search, implemented in the typical, iterative fashion;
- a parallelized BFS which makes use of Go's channels and goroutines.

## Dependencies
- [gographviz](https://github.com/awalterschulze/gographviz)

#TODO: installation instructions for dependencies


## Usage
Navigate to the project's root and build it with
```
go build
```

The program can be started as 
```
./graphxplorer [-verbose] <path>
```
where `path` is the path to a `.gv` file. The `example_inputs` containes a couple of example files that can be used to test the program.

The `verbose` flag can be used to obtain more details about the search.

## Communication Graph
#TODO