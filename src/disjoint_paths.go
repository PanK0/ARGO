package main

import "fmt"

/*
	A disjoint path solution for a node A is a set of paths from A to any other generic node
*/

// Dictionnary with key the ID of a node and value a list of paths
// The paths are represented as a list of strings
type DisjointPaths struct {
	paths map[string] [][]string
}

// return a new DisjointPaths
func NewDisjointPaths() *DisjointPaths {
	return &DisjointPaths{
		paths: make(map[string] [][]string, 0),
	}
}

// Add a path to the disjoint path
func (dp *DisjointPaths) Add(path []string, node_id string) {
	if _, ok := dp.paths[node_id]; !ok {
		dp.paths[node_id] = make([][]string, 0)
	}
	dp.paths[node_id] = append(dp.paths[node_id], path)
}

// Delete a path from the disjoint path
func (dp *DisjointPaths) deleteElement(node_id string) {
	delete(dp.paths, node_id)
}

// Get the paths for a given node
func (dp *DisjointPaths) Get(node_id string) [][]string {
	return dp.paths[node_id]
}

// Get the paths for all nodes
func (dp *DisjointPaths) GetAll() map[string] [][]string {
	return dp.paths
}

// Get the number of paths for a given node
func (dp *DisjointPaths) GetNumberOfPaths(node_id string) int {
	return len(dp.paths[node_id])
}

// Get the number of paths for all nodes
func (dp *DisjointPaths) GetNumberOfPathsAll() map[string] int {
	numberOfPaths := make(map[string] int, 0)
	for k, v := range dp.paths {
		numberOfPaths[k] = len(v)
	}
	return numberOfPaths
}

// Transform the DisjointPaths into a string
func (dp *DisjointPaths) toString() string {
	msg := ""
	for k, v := range dp.paths {
		msg += fmt.Sprintf("%s - %s\n", k, v)
	}
	return msg
}

// Given a DisjointPaths key, transform the key and the values into a string that can be lately be reinserted into the DisjointPaths
func (dp *DisjointPaths) dp_toString(key string) string {
	msg := ""
	for _, v := range dp.paths[key] {
		msg += fmt.Sprintf("%s - ", v)
	}
	return msg
}