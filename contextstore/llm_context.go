// contextstore/llm_context.go
package contextstore

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// LLMRequest defines the request payload for Ollama API
type LLMRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// LLMStreamChunk represents one chunk in Ollama's stream response
type LLMStreamChunk struct {
	Response string `json:"response"`
	Final    bool   `json:"done"`
}

func QueryLLM(model, prompt string, stream bool) string {
	return QueryLLMWithSession("", model, prompt, stream)
}

func QueryLLMWithSession(sessionID, model, prompt string, stream bool) string {
	reqBody := LLMRequest{
		Model:  model,
		Prompt: prompt,
		Stream: stream,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Sprintf("[error: encode request: %v]", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:11434/api/generate", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Sprintf("[error: build request: %v]", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Sprintf("[error: send request: %v]", err)
	}
	defer resp.Body.Close()

	var output string

	if !stream {
		var result LLMStreamChunk
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return fmt.Sprintf("[error: decode response: %v]", err)
		}
		output = result.Response
	} else {
		var sb strings.Builder
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			var chunk LLMStreamChunk
			if err := json.Unmarshal([]byte(line), &chunk); err == nil {
				sb.WriteString(chunk.Response)
			}
		}
		output = sb.String()
	}

	if sessionID != "" {
		AppendToMemory(sessionID, prompt+"\n"+output)
	}

	return output
}
