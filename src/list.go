package main

import (
	"strings"
)

/*
	START OPERATIONS ON LIST STRUCT
*/

// Create a list of strings
type List struct {
	objects []string
}

// Return a new List
func NewList() *List {
	return &List{
		objects: make([]string, 0),
	}
}

// Add an element to a List
func (l *List) Add(str string) {
	l.objects = append(l.objects, str)
}

// Get all the elements of the List
func (l *List) GetAll() []string {
	return l.objects
}

/*
	END OPERATIONS ON LIST STRUCT
*/

// Find an element in an array and return it with its position
func findElement(arr []string, element string) (string, int) {
	for i, v := range arr {
		if v == element {
			return v, i
		}
	}
	return "", -1
}

// Extract a substring inside apici that should be the text of the message
func extractMessage(input string) string {
    start := strings.Index(input, "\"") + 1
    end := strings.LastIndex(input, "\"")
    return input[start:end]
}

// Given a full address, returns a string ID
func getNodeID(addr string) string {
	parts := strings.Split(addr, "/")
	return parts[len(parts)-1]
}

// Check wether a string is into an array of strings
func contains(arr []string, addr_id string) bool {
    for _, a := range arr {
		a_id := getNodeID(a)
        if a_id == addr_id {
            return true
        }
    }
    return false
}

// Given two lists of strings, check whether they contain the same elements
// return 0 if they contain the same elements
// return -1 if they don't
//lint:ignore U1000 Unused function for future use
func compareLists(a []string, b []string) int {
	if len(a) != len(b) {
		return -1
	}
	for _, e := range a {
		if !contains(b, getNodeID(e)) {
			return -1
		}
	}
	return 0
}

// Given two lists of strings, checke whether one is contained in the other
// return 0 if they contain the same elements
// return -1 if they don't
//lint:ignore U1000 Unused function for future use
func isSubSet(a []string, b []string) int {
	if len(a) > len(b) {
		return -1
	}
	for _, e := range a {
		if !contains(b, getNodeID(e)) {
			return -1
		}
	}
	return 0
}

// Given a list of lists of strings, transform it into a string in json format that can be later converted into a list of lists of strings
//lint:ignore U1000 Unused function for future use
func convertListToString(lists [][]string) string {
	var result string
	for _, list := range lists {
		result += "["
		for i, str := range list {
			result += "\"" + str + "\""
			if i != len(list)-1 {
				result += ", "
			}
		}
		result += "]"
	}
	return result
}

// Given a string in json format, convert it into a list of lists of strings
//lint:ignore U1000 Unused function for future use
func convertStringToList(str string) [][]string {
	var result [][]string
	str = strings.Trim(str, "[]")
	if str == "" {
		return result
	}
	lists := strings.Split(str, "], [")
	for _, list := range lists {
		list = strings.Trim(list, "[]")
		if list == "" {
			continue
		}
		elements := strings.Split(list, ", ")
		result = append(result, elements)
	}
	return result
} 