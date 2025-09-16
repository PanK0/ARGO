package main

import (
	"math/rand"
	"time"
)

/*
	CTOP
	confirmed topology
	key: process ID
	value: processes' neighbourhood
*/

// cTop.toString() method moved to output_print_functions.go

type CTop struct {
	tuples map[string] []string 
}

// Return a new cTop
func NewCTop() *CTop {
	return &CTop {
		tuples: make(map[string] []string, 0),
	}
}

// DeepCopy creates a deep copy of the CTop object
func (c *CTop) DeepCopy() *CTop {
	temp := CTop{
		tuples : make(map[string][]string, len(c.tuples)),
	}

	// Copy each map entry (deep copy of maps)
	for key, neighbours := range c.tuples {
		newNeighbours := make([]string, len(neighbours))
		copy(newNeighbours, neighbours) // Copy the slice contents
		temp.tuples[key] = newNeighbours
	}

	return &temp
}

// RemoveElement removes an element from the CTop by its key.
func (c *CTop) RemoveElement(key string) {
    delete(c.tuples, key)
}

// TotalRemoveElement removes every entry of node_id both as a key and as any occurrence in the values of any key.
func (c *CTop) TotalRemoveElement(node_id string) {
    // Remove node_id as a key
    c.RemoveElement(node_id)

    // Remove node_id from all neighbourhoods (values)
    for key, neighbours := range c.tuples {
        newNeighbours := make([]string, 0, len(neighbours))
        for _, n := range neighbours {
            if n != node_id {
                newNeighbours = append(newNeighbours, n)
            }
        }
        c.tuples[key] = newNeighbours
    }
}

// Add a node neighbour to the neighbourhood of node node
func (ctop CTop) AddNeighbour(node string, neighbour string) {
	// check whether the neighbour is already in the neighbourhood
	for _, n := range ctop.tuples[node] {
		if n == neighbour {
			return
		}
	}
	ctop.tuples[node] = append(ctop.tuples[node], neighbour)
}

// Add a neighbourhood to the neighbourhood of node node
func (top CTop) AddNeighbourhood(node string, neighbours []string) {
	// substitute the neighbourhood of the node with the new one
	top.tuples[node] = neighbours	
}

// Get all the neighbourhood of a node
func (top CTop) GetNeighbourhood(node string) []string {
	return top.tuples[node]
}

// Given a node, check whether there is 
// some node's information in topology's cTop
func (ctop CTop) checkInCTop(node string) bool {
	for k := range ctop.tuples {
		if (k == node) {
			return true
		}
	}
	return false
}

// ConvertGraphToCTop converts a Graph into a CTop struct
func ConvertGraphToCTop(graph *Graph) *CTop {
    ctop := NewCTop()

    for node, neighbors := range graph.adjList {
        for _, neighbor := range neighbors {
            ctop.AddNeighbour(node, neighbor)
        }
    }

    return ctop
}

// Convert a CTop into a Graph
func ConvertCTopToGraph(ctop *CTop) *Graph {
	graph := NewGraph()

	for node, neighbours := range ctop.tuples {
		for _, neighbour := range neighbours {
			graph.AddEdge(node, neighbour)
		}
	}

	return graph
}

// CTop to Graph conversion with Byzantine fault tolerance
// Ref @ `Tractable Reliable Communication in Compromised Networks, Giovanni Farina` - cpt. 9.4 - Explorer2 - pg 78 
func exp2_ConvertCTopToGraph(top *Topology, autoRec bool) *Graph {
    graph := NewGraph()
    vertices := make(map[string]bool)

	// if autoRec is true, we load the neighbourhood of this node
	if autoRec {
		// Load the neighbourhood of the current node
		for _, neighbour := range top.ctop.GetNeighbourhood(top.nodeID) {
			vertices[neighbour] = true
		}
	}

	// Rule 1: ∀<u, Γ(u)> ∈ cTopi ⇒ ∃u ∈ Vi
	// Rule 1: A node u is inserted in Vi if the related entry is in cTopi
    for node := range top.ctop.tuples {
        vertices[node] = true
    }

	// Rule 2: . ∀v ∈ Γ(u),<u, Γ(u)> ∈ cTop : X ← U u, |X| > f ⇒ ∃v ∈ Vi
	// Rule 2: A node v is inserted in Vi if v is declared as neighbour by at least f+1 different nodes
    for _, neighbours := range top.ctop.tuples {
        for _, neighbour := range neighbours {
            count := 0
            for _, otherNeighbours := range top.ctop.tuples {
                if isInNeighbourhood(neighbour, otherNeighbours) {
                    count++
                }
            }
            if count > MAX_BYZANTINES {
                vertices[neighbour] = true
            }
        }
    }

	// Rule 3: ∀<v, Γ(v)> ∈ cTopi, u ∈ Γ(v), u ∈ Vi ⇒ ∃(v, u) ∈ Ei
	// Rule 3: An edge (v, u) is added in Ei if both nodes are in Vi and v declares u in its neighbourhood
    for node, neighbours := range top.ctop.tuples {
        for _, neighbour := range neighbours {
            if vertices[node] && vertices[neighbour] {
                graph.AddEdge(node, neighbour)
            }
        }
    }

    return graph
}

// Convert CTop to Graph
func generateGraph(top *Topology, autoRecognizeNeighbours bool) *Graph {
	return exp2_ConvertCTopToGraph(top, autoRecognizeNeighbours)
	
}

// Replaces the current cTop with a new one by loading a new graph
// by deleting the old one
func loadCTop(graph *Graph) *CTop {
	top := ConvertGraphToCTop(graph)
	return top 
}

// Only load neighbourhood
func (ctop CTop) loadNeigh(graph *Graph, thisNode string) {
	top := ConvertGraphToCTop(graph)
	ctop.tuples[thisNode] = top.tuples[thisNode]
}

// Get all nodes in the cTop
func (ctop CTop) GetAllNodes() []string {
	nodes := make([]string, 0, len(ctop.tuples))
	for node := range ctop.tuples {
		nodes = append(nodes, node)
	}
	return nodes
}



/*
	UTOP
	uncomnfirmed topology
	key: process ID
	value: processes' neighbourhood, visited set
*/
type UTop struct {
	tuples map[string] [2][]string
}

// Return a new uTop
func NewUTop() *UTop {
	return &UTop {
		tuples: make(map[string] [2][]string, 0),
	}
}

// Add a node neighbour to the neighbourhood of node node
func (utop UTop) AddNeighbour(node string, neighbour string) {
	// Avoid a node being neighbour with itself
	if node == neighbour {
		return
	}
	// Check whether the neighbour is already in the neighbourhood
	for _, n := range utop.tuples[node][0] {
		if n == neighbour {
			return
		}
	}
	temp := utop.tuples[node]
	temp[0] = append(temp[0], neighbour)
	utop.tuples[node] = temp
}

// Add a neighbourhood to the neighbourhood of node node
func (utop UTop) AddNeighbourhood(node string, neighbours []string) {
	for _, n := range neighbours {
		utop.AddNeighbour(node, n)
	}
}

// RemoveElement removes an element from the CTop by its key.
func (u *UTop) RemoveElement(key string) {
    delete(u.tuples, key)
}

// Add a node's id to the visited set of the node node
func (utop UTop) AddVisited(node string, visited string) {
	// Check whether the visited is already in the visited set
	for _, n := range utop.tuples[node][1] {
		if n == visited {
			return
		}
	}
	temp := utop.tuples[node]
	temp[1] = append(temp[0], visited)
	utop.tuples[node] = temp
}

// Add a visited set to a  node
func (utop UTop) AddVisitedSet(node string, visitedSet []string) {
	for _, v := range visitedSet {
		utop.AddVisited(node, v)
	}
}

// Add a node to uTop with its neighbourhood and visited set
func (utop UTop) AddElement(node string, neighbourhood, visitedSet []string) {
	// substitute existent information with the new one
	utop.AddNeighbourhood(node, neighbourhood)
	utop.AddVisitedSet(node, visitedSet)
}

// Given a string node and a []string neighbourhood, 
// checks wether node is in the neighbourhood
func isInNeighbourhood(node string, list []string) bool {
	for _, n := range list {
		if node == n {
			return true
		}
	}
	return false
}

// Given a tuple (node, neighbouhood), check wether there is
// a tuple in uTop with the same combination id, neighbourhood
func (utop UTop) checkInUTopNeigh(node string, neighbourhood []string) bool {
	
	if len(utop.tuples[node][0]) != len(neighbourhood) {
		return false
	}
	
	for _, n := range neighbourhood {
		if !isInNeighbourhood(n, utop.tuples[node][0]) {
			return false
		}
	}

	return true
}

func (utop UTop) checkInUTop(node string) bool {
	_, exists := utop.tuples[node]
	return exists
}

// Get all the neighbourhood of a node
func (utop UTop) GetNeighbourhood(node string) []string {
	return utop.tuples[node][0]
}


// ConvertGraphToUTop converts a Graph into a UTop struct
//lint:ignore U1000 Unused function for future use
func ConvertGraphToUTop(graph *Graph) *UTop {
	utop := NewUTop()

	for node, neighbors := range graph.adjList {
		for _, neighbor := range neighbors {
			utop.AddNeighbour(node, neighbor)
		}
	}

	return utop
}


/*
	TOPOLOGY
	Contains both cTop and uTop
	It is proper of a node
*/
type Topology struct {
	nodeID string
	ctop CTop
	utop UTop
}

// Return a new topology struct
func NewTopology() *Topology {
	return &Topology{
		nodeID: "",
		ctop: 	*NewCTop(),
		utop:	*NewUTop(),
	}
}

// Get cTop and uTop
func (top Topology) GetCTop() CTop {
	return top.ctop
}

func (top Topology) GetUTop() UTop {
	return top.utop
}

// Reset the topology by deleting both cTop and uTop
func (top *Topology) Reset() {
	top.ctop = *NewCTop()
	top.utop = *NewUTop()
}


// Computes an intersection betweens two []string objects
func Intersect(a []string, b []string) []string {
    m := make(map[string]bool)
    for _, item := range a {
        m[item] = true
    }
    var result []string
    for _, item := range b {
        if m[item] {
            result = append(result, item)
        }
    }
    return result
}

// GetRandomNeighbour returns a random neighbour of this node from the topology.
// Returns an empty string if the node has no neighbours.
func (top *Topology) GetRandomNeighbour() string {
    neighbours := top.ctop.GetNeighbourhood(top.nodeID)
    if len(neighbours) == 0 {
        return ""
    }
    rand.Seed(time.Now().UnixNano())
    idx := rand.Intn(len(neighbours))
    return neighbours[idx]
}