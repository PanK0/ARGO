// This module is a graph implementation with the Ford-Fulkerson algorithm to find the maximum flow in a network flow problem.
// It can be useful during the network topoly discovery, where the max flow can be used to determine the number of disjoint paths
// between two nodes in a network.

package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

// Graph represents a graph where nodes are identified by strings
type Graph struct {
	adjList map[string][]string // Adjacency list to store edges
	nodes   map[string]bool     // Set of all nodes in the graph
}

// NewGraph creates a new graph
func NewGraph() *Graph {
	return &Graph{
		adjList: make(map[string][]string),
		nodes:   make(map[string]bool),
	}
}


// AddEdge adds an edge between two nodes (bidirectional edges for residual capacity).
// If the nodes are already connected, it does nothing.
func (g *Graph) AddEdge(from, to string) {
	if !g.isEdgePresent(from, to) {
		g.adjList[from] = append(g.adjList[from], to)
		g.adjList[to] = append(g.adjList[to], from) // Add reverse edge for bidirectional graph
		g.nodes[from] = true
		g.nodes[to] = true
	}
}

// isEdgePresent checks if an edge between two nodes exists
func (g *Graph) isEdgePresent(from, to string) bool {
	for _, neighbor := range g.adjList[from] {
		if neighbor == to {
			return true
		}
	}
	return false
}

// RemoveEdge removes an edge between two nodes
func (g *Graph) RemoveEdge(from, to string) {
	g.adjList[from] = removeFromSlice(g.adjList[from], to)
	g.adjList[to] = removeFromSlice(g.adjList[to], from)
}

// ModifyEdge modifies an edge by removing the old edge and adding a new one
func (g *Graph) ModifyEdge(fromOld, toOld, fromNew, toNew string) {
	g.RemoveEdge(fromOld, toOld)
	g.AddEdge(fromNew, toNew)
}

// GetNeighbors returns the neighbors of a given node
func (g *Graph) GetNeighbors(node string) []string {
	return g.adjList[node]
}

// DFS performs a Depth-First Search to find an augmenting path
// Used in Ford-Fulkerson algorithm
func (g *Graph) DFS(residualGraph map[string]map[string]bool, source, sink string, parent map[string]string) bool {
	visited := make(map[string]bool)
	stack := []string{source}
	visited[source] = true

	for len(stack) > 0 {
		u := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		for _, v := range g.adjList[u] {
			// If there is residual capacity and v is not visited
			if !visited[v] && residualGraph[u][v] {
				parent[v] = u
				if v == sink {
					return true // Found a path to the sink
				}
				stack = append(stack, v)
				visited[v] = true
			}
		}
	}

	return false
}

// FordFulkerson computes the maximum flow using the Ford-Fulkerson algorithm
func (g *Graph) FordFulkerson(source, sink string) int {
	// Create a residual graph initialized with true (1 capacity for all edges)
	residualGraph := make(map[string]map[string]bool)
	for u := range g.nodes {
		residualGraph[u] = make(map[string]bool)
		for _, v := range g.adjList[u] {
			residualGraph[u][v] = true // Edge exists initially
		}
	}

	parent := make(map[string]string) // To store the path
	maxFlow := 0                      // Initialize max flow to 0

	// Augment the flow while there is an augmenting path
	for g.DFS(residualGraph, source, sink, parent) {
		// Since each edge has capacity 1, the bottleneck is always 1
		// Update residual capacities along the path
		v := sink
		for v != source {
			u := parent[v]
			residualGraph[u][v] = false // Forward edge is used
			residualGraph[v][u] = true  // Reverse edge is added
			v = u
		}

		// Increase the flow by 1
		maxFlow++
	}

	return maxFlow
}

// RemoveNode temporarily removes a node from the graph
func (g *Graph) RemoveNode(node string) map[string][]string {
	backup := make(map[string][]string)
	backup[node] = g.adjList[node] // Backup the node's neighbors

	// Remove node's edges from its neighbors
	for _, neighbor := range g.adjList[node] {
		newNeighbors := []string{}
		for _, n := range g.adjList[neighbor] {
			if n != node {
				newNeighbors = append(newNeighbors, n)
			}
		}
		g.adjList[neighbor] = newNeighbors
	}

	// Remove the node
	g.adjList[node] = nil

	return backup
}

// RestoreNode restores a node and its edges into the graph
func (g *Graph) RestoreNode(node string, backup map[string][]string) {
	g.adjList[node] = backup[node]

	// Restore the edges to the neighbors
	for _, neighbor := range g.adjList[node] {
		g.adjList[neighbor] = append(g.adjList[neighbor], node)
	}
}

// NodeConnectivity computes the node connectivity of the graph
func (g *Graph) nodeConnectivity() int {
	min_connectivity := len(g.nodes)
	for node_1 := range g.nodes {
		for node_2 := range g.nodes {
			if node_1 != node_2 {
				max_conn := g.FordFulkerson(node_1, node_2)
				if max_conn < min_connectivity {	
					min_connectivity = max_conn
				}
			}
		}
	}
	return min_connectivity
}


// Helper function to remove a node from a slice
func removeFromSlice(slice []string, value string) []string {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}


// LoadGraphFromCSV reads a CSV file and constructs the graph
func LoadGraphFromCSV(filePath string) *Graph {
    file, err := os.Open(filePath)
    if err != nil {
        log.Fatalf("Failed to open file: %v", err)
    }
    defer file.Close()

    graph := NewGraph()
    reader := csv.NewReader(bufio.NewReader(file))
    lines, err := reader.ReadAll()
    if err != nil {
        log.Fatalf("Failed to read CSV file: %v", err)
    }

    // Skip the header and parse the remaining rows
    for _, line := range lines[1:] {
        node := line[0]
        var neighbors []string
        for _, cell := range line[1:] {
            if cell != "" {
                neighbors = append(neighbors, cell)
            }
        }
        for _, neighbor := range neighbors {
            graph.AddEdge(node, neighbor)
        }
    }

    return graph
}

// Returns a string
func (g *Graph) GraphToString() string {
	str := "Graph: "
	for node, neighbors := range g.adjList {
		// Get the last 5 characters of the node
		nodeToPrint := addressToPrint(node, NODE_PRINTLAST)
		// Print the node and its neighbors
		str += fmt.Sprintf("%s -> [", nodeToPrint)
		for i, neighbor := range neighbors {
			// Get the last 5 characters of the neighbor
			neighborToPrint := addressToPrint(neighbor, NODE_PRINTLAST)
			if i > 0 {
				str += ", "
			}
			str += neighborToPrint
		}
		str += "]; "
	}
	return str
}