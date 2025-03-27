// cmd/server/rest_server.go
package main

import (
	"agentd/proto"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/grpc"
)

type jsonRequest struct {
	SessionID   string            `json:"session_id"`
	Instruction string            `json:"instruction"`
	Metadata    map[string]string `json:"metadata"`
}

type jsonResponse struct {
	Content  string            `json:"content"`
	Trace    map[string]string `json:"trace"`
	Complete bool              `json:"complete"`
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h(w, r)
	}
}

func StartRestGateway() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := proto.NewInstructionServiceClient(conn)

	http.HandleFunc("/v1/execute", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Use POST", http.StatusMethodNotAllowed)
			return
		}

		var req jsonRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		resp, err := client.SendInstruction(context.Background(), &proto.InstructionRequest{
			SessionId:   req.SessionID,
			Instruction: req.Instruction,
			Metadata:    req.Metadata,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("gRPC error: %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Printf("Resp: %s \n", resp.Content)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jsonResponse{
			Content:  resp.Content,
			Trace:    resp.Trace,
			Complete: resp.Complete,
		})
	}))

	fmt.Println("🌐 REST Gateway available at http://localhost:9090/v1/execute")
	go http.ListenAndServe(":9090", nil)
}
