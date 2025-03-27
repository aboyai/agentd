package router

import (
	"fmt"
	"sort"
	"strings"
)

type DAG struct {
	Nodes    map[string]*DAGNode
	Children map[string][]string // parentID → childIDs
}

type DAGNode struct {
	ID         string
	Op         string
	Next       []string
	Conditions map[string]string // parentID → condition (e.g., ?contains:heads)
}

func NewDAGFromPlan(plan string) *DAG {
	dag := &DAG{
		Nodes:    map[string]*DAGNode{},
		Children: map[string][]string{},
	}

	plan = strings.TrimPrefix(plan, "plan://")
	parts := strings.Split(plan, ";")

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		if strings.Contains(p, "=") && !strings.Contains(p, "<-") {
			kv := strings.SplitN(p, "=", 2)
			id := strings.TrimSpace(kv[0])
			op := strings.TrimSpace(kv[1])
			dag.Nodes[id] = &DAGNode{ID: id, Op: op, Next: []string{}, Conditions: map[string]string{}}
		} else if strings.Contains(p, "<-") {
			kv := strings.SplitN(p, "<-", 2)
			child := strings.TrimSpace(kv[0])
			parentCond := strings.TrimSpace(kv[1])

			parent := parentCond
			cond := ""
			if strings.Contains(parentCond, "?") {
				split := strings.SplitN(parentCond, "?", 2)
				parent = strings.TrimSpace(split[0])
				cond = strings.TrimSpace(split[1])
			}

			// Ensure both parent and child nodes exist
			if dag.Nodes[parent] == nil {
				dag.Nodes[parent] = &DAGNode{ID: parent, Next: []string{}, Conditions: map[string]string{}}
			}
			if dag.Nodes[child] == nil {
				dag.Nodes[child] = &DAGNode{ID: child, Next: []string{}, Conditions: map[string]string{}}
			}

			dag.Nodes[parent].Next = append(dag.Nodes[parent].Next, child)
			dag.Children[parent] = append(dag.Children[parent], child)
			if cond != "" {
				dag.Nodes[child].Conditions[parent] = cond
			}
		}
	}
	return dag
}

func (dag *DAG) Execute(sessionID string, meta map[string]string) (string, map[string]string) {
	visited := map[string]bool{}
	trace := map[string]string{}
	results := map[string]string{}
	pending := map[string]*DAGNode{}

	// queue of ready-to-execute nodes
	queue := []*DAGNode{}

	// Stage 1: find root nodes (no conditions or parents)
	for _, node := range dag.Nodes {
		if len(node.Conditions) == 0 {
			queue = append(queue, node)
		} else {
			pending[node.ID] = node
		}
	}

	// Stage 2: execute ready queue, re-evaluate pending when parents are done
	for len(queue) > 0 {
		nextQueue := []*DAGNode{}

		for _, node := range queue {
			skip := false
			hasCondition := len(node.Conditions) > 0

			// Evaluate conditions
			for parentID, cond := range node.Conditions {
				if !visited[parentID] {
					skip = true
					break
				}
				if cond == "else" {
					continue // handled later
				}
				parentOutput := results[parentID]
				if !evaluateCondition(parentOutput, cond) {
					skip = true
					break
				}
			}

			if hasCondition && skip {
				trace[node.ID] = "Skipped due to condition"
				visited[node.ID] = true
				continue
			}

			// 🧠 Execute node
			fmt.Printf("🔁 Executing node %s = %s\n", node.ID, node.Op)
			out, stepTrace := Dispatch(sessionID, node.Op, meta)
			trace[node.ID] = fmt.Sprintf("%s => %s", node.Op, out)
			for k, v := range stepTrace {
				trace[node.ID+"."+k] = v
			}
			results[node.ID] = out
			visited[node.ID] = true

			// Queue next nodes
			for _, nextID := range node.Next {
				if visited[nextID] {
					continue
				}
				allParentsReady := true
				for parent := range dag.Nodes[nextID].Conditions {
					if !visited[parent] {
						allParentsReady = false
						break
					}
				}
				if allParentsReady {
					nextQueue = append(nextQueue, dag.Nodes[nextID])
					delete(pending, nextID)
				}
			}
		}

		queue = nextQueue
	}

	// Stage 3: handle ?else
	// Handle all pending ?else nodes last
	for _, node := range pending {
		for parentID, cond := range node.Conditions {
			if cond != "else" {
				continue
			}

			// Check if any sibling of this node (except self) executed successfully
			shouldSkipElse := false
			for _, siblingID := range dag.Children[parentID] {
				if siblingID == node.ID {
					continue
				}
				if val, ok := trace[siblingID]; ok && !strings.HasPrefix(val, "Skipped") {
					shouldSkipElse = true
					break
				}
			}

			if shouldSkipElse {
				trace[node.ID] = "Skipped else — sibling matched"
				visited[node.ID] = true
				break
			}

			// No sibling executed — proceed
			fmt.Printf("🛟 Executing fallback node %s = %s\n", node.ID, node.Op)
			output, stepTrace := Dispatch(sessionID, node.Op, meta)
			trace[node.ID] = fmt.Sprintf("%s => %s", node.Op, output)
			for k, v := range stepTrace {
				trace[node.ID+"."+k] = v
			}
			results[node.ID] = output
			visited[node.ID] = true
		}
	}

	// Return final leaf result
	final := ""
	for id, node := range dag.Nodes {
		if len(node.Next) == 0 && visited[id] && !strings.HasPrefix(trace[id], "Skipped") {
			final = id
			break
		}
	}
	return results[final], trace
}

func evaluateCondition(output, condition string) bool {
	output = strings.ToLower(strings.TrimSpace(output))
	if idx := strings.Index(output, "=>"); idx != -1 {
		output = strings.TrimSpace(output[idx+2:])
	}
	if strings.HasPrefix(condition, "contains:") {
		values := strings.Split(strings.TrimPrefix(condition, "contains:"), "|")
		words := strings.Fields(output)
		for _, keyword := range values {
			keyword = strings.ToLower(strings.TrimSpace(keyword))
			for _, word := range words {
				if word == keyword {
					return true
				}
			}
		}
		return false
	}
	return true
}

func (dag *DAG) TopologicalSort() []string {
	visited := map[string]bool{}
	order := []string{}

	var visit func(string)
	visit = func(n string) {
		if visited[n] {
			return
		}
		visited[n] = true
		for _, m := range dag.Nodes[n].Next {
			visit(m)
		}
		order = append(order, n)
	}

	keys := make([]string, 0, len(dag.Nodes))
	for k := range dag.Nodes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		visit(k)
	}
	return order
}

func (dag *DAG) ToMermaidWithTrace(trace map[string]string) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")
	for id, node := range dag.Nodes {
		label := id
		if node.Op != "" {
			label += "[" + strings.ReplaceAll(node.Op, ":", ": ") + "]"
		}
		sb.WriteString(fmt.Sprintf("%s\n", label))
		for _, next := range node.Next {
			sb.WriteString(fmt.Sprintf("%s --> %s\n", id, next))
		}
	}
	for id, val := range trace {
		if strings.HasPrefix(val, "Skipped") {
			sb.WriteString(fmt.Sprintf("style %s fill:#333,color:#999,stroke:#999\n", id))
		}
	}
	return sb.String()
}

func (dag *DAG) ExportTrace(trace map[string]string) string {
	var sb strings.Builder
	for id, val := range trace {
		sb.WriteString(fmt.Sprintf("%s => %s\n", id, val))
	}
	return sb.String()
}
