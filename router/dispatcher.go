// router/dispatcher.go
package router

import (
	"agentd/contextstore"
	"agentd/tools"
	"fmt"
	"strings"
)

func Dispatch(sessionID, instruction string, meta map[string]string) (string, map[string]string) {
	trace := make(map[string]string)

	switch {
	case strings.HasPrefix(instruction, "llm://"):
		model, prompt := parseLLM(instruction)
		trace["type"] = "llm"
		trace["raw_model"] = model

		if model == "" {
			trace["error"] = "missing model"
			return "Unsupported instruction", trace
		}

		useMemory := false
		if strings.Contains(model, ".memory") {
			useMemory = true
			model = strings.Split(model, ".")[0]
			trace["context"] = "memory"
		}

		trace["model"] = model

		if useMemory {
			memory, _ := contextstore.FetchMemory(sessionID)
			prompt = memory + "\n" + prompt
		}

		output := contextstore.QueryLLMWithSession(sessionID, model, prompt, false)
		return output, trace

	case strings.HasPrefix(instruction, "tool://"):
		toolName := strings.TrimPrefix(instruction, "tool://")
		trace["type"] = "tool"

		if strings.HasPrefix(toolName, "search:") {
			query := strings.TrimPrefix(toolName, "search:")
			result := tools.WebSearch(query)
			trace["tool"] = "web_search"
			trace["query"] = query
			return result, trace
		}

		trace["tool"] = toolName
		return fmt.Sprintf("[Tool: %s executed]", toolName), trace

	case strings.HasPrefix(instruction, "plan://"):
		dag := NewDAGFromPlan(instruction)
		trace["type"] = "plan"
		out, planTrace := dag.Execute(sessionID, meta)
		for k, v := range planTrace {
			trace[k] = v
		}
		return out, trace

	default:
		trace["type"] = "unknown"
		trace["error"] = instruction
		return "Unsupported instruction", trace
	}
}

func parseLLM(uri string) (model string, prompt string) {
	parts := strings.SplitN(strings.TrimPrefix(uri, "llm://"), ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	} else if len(parts) == 1 {
		return parts[0], ""
	}
	return "default", uri
}
