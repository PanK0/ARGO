package main

import "fmt"

type Message struct {
	ID				string			`json:"id"`
	Type 			string			`json:"type"`
	Sender 			string			`json:"sender"`
	Source 			string			`json:"source"`
	Target 			string			`json:"target"`
	Content			string			`json:"content"`
	Neighbourhood 	[]string		`json:"neighbourhood"`
	Path			[]string 		`json:"path"`
}

func msgToString(m Message) string {
	id := fmt.Sprintf("MSG ID:		%s\n", m.ID)
	ty := fmt.Sprintf("TYPE:		%s\n", m.Type)
	se := fmt.Sprintf("SENDER:		%s\n", m.Sender)
	so := fmt.Sprintf("SOURCE:		%s\n", m.Source)
	ta := fmt.Sprintf("TARGET:		%s\n", m.Target)
	co := fmt.Sprintf("CONTENT:	%s\n", m.Content)
	ne := "NEIGHBOURHOOD:	\n"
	pa := "PATH:		\n"
	
	for i, p := range m.Neighbourhood {
		ne += fmt.Sprintf("	%d - %s\n", i, p)
	}
	
	for i, p := range m.Path {
		pa += fmt.Sprintf("	%d - %s\n", i, p)
	}

	switch m.Type {
	case TYPE_DIRECT_MSG : 
		ne = ""
		pa = ""
	case TYPE_BROADCAST :
		ne = ""
	case TYPE_DETECTOR :
		se = ""
		ta = ""
		co = ""
		pa = ""
	case TYPE_EXPLORER :
		ta = ""
		co = ""
	case TYPE_EXPLORER2 :
		ta = ""
	case TYPE_CRC_EXP :
		ta = ""
	case TYPE_CRC_ROU :
		co = "" 
		ne = ""
	case TYPE_CRC_CNT :
		ne = ""
	}

	msg := id + ty + se + so + ta + co + ne + pa
	
	return msg
}

// Compare two Message objects for equality
func equalMessage(a, b Message) bool {
    if a.ID != b.ID ||
        a.Type != b.Type ||
        a.Sender != b.Sender ||
        a.Source != b.Source ||
        a.Target != b.Target ||
        a.Content != b.Content {
        return false
    }
    if len(a.Neighbourhood) != len(b.Neighbourhood) {
        return false
    }
    for i := range a.Neighbourhood {
        if a.Neighbourhood[i] != b.Neighbourhood[i] {
            return false
        }
    }
    if len(a.Path) != len(b.Path) {
        return false
    }
    for i := range a.Path {
        if a.Path[i] != b.Path[i] {
            return false
        }
    }
    return true
}