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
func (dp *DisjointPaths) Add(node_id string, path []string) {
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

// Check if a path is already present in the DisjointPaths
func (dp *DisjointPaths) containsPath(node_id string, path []string) bool {
	for _, p := range dp.paths[node_id] {
		if len(p) != len(path) {
			continue
		}
		equal := true
		for i := 0; i < len(p); i++ {
			if p[i] != path[i] {
				equal = false
				break
			}
		}
		if equal {
			return true
		}
	}
	return false
}

// Given a DisjointPaths object, merge it with another one by adding the paths of the second one to the first one if the paths are not already present
func (dp *DisjointPaths) MergeDP(dp2 *DisjointPaths) {
	for k, v := range dp2.paths {
		if _, ok := dp.paths[k]; !ok {
			dp.paths[k] = make([][]string, 0)
		}
		for _, path := range v {
			if !dp.containsPath(k, path) {
				dp.paths[k] = append(dp.paths[k], path)
			}
		}
	}
}


// Print DisjointPaths
func (dp *DisjointPaths) toString() string {
	str := "Disjoint Paths:"
	for node, paths := range dp.paths {
		nodetoprint := addressToPrint(node, NODE_PRINTLAST)
		str += fmt.Sprintf("Node %s :\n", nodetoprint)
		for i, path := range paths {
			str += fmt.Sprintf("\tPath %d : [", i)
			for _, p := range path {
				ptoprint := addressToPrint(p, NODE_PRINTLAST)
				str += fmt.Sprintf(" %s , ", ptoprint)
			}
			str += fmt.Sprintf(" ]\n")
		}
		str += fmt.Sprint("\n")
	}
	return str
}

// Given a DisjointPaths key, transform the key and the values into a string that can be lately be reinserted into the DisjointPaths
func (dp *DisjointPaths) dp_toString(key string) string {
	msg := ""
	for _, v := range dp.paths[key] {
		msg += fmt.Sprintf("%s - ", v)
	}
	return msg
}