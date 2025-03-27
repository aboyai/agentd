// router/mermaid.go
package router

import (
	"fmt"
	"strings"
)

// MermaidString returns the DAG in Mermaid.js flowchart format
func (d *DAG) MermaidString() string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	for _, node := range d.Nodes {
		sb.WriteString(fmt.Sprintf("  %s[%q]\n", node.ID, node.Op))
		for _, child := range node.Next {
			sb.WriteString(fmt.Sprintf("  %s --> %s\n", node.ID, child))
		}
	}

	return sb.String()
}
