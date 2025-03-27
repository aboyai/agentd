// api/handler.go
package api

import (
	"agentd/proto"
	"agentd/router"
	"context"
)

type AgentDHandler struct {
	proto.UnimplementedInstructionServiceServer
}

func (h *AgentDHandler) SendInstruction(ctx context.Context, req *proto.InstructionRequest) (*proto.InstructionResponse, error) {
	meta := req.GetMetadata()
	result, trace := router.Dispatch(req.GetSessionId(), req.GetInstruction(), meta)
	return &proto.InstructionResponse{
		Content:  result,
		Trace:    trace,
		Complete: true,
	}, nil
}

func (h *AgentDHandler) StreamInstruction(req *proto.InstructionRequest, stream proto.InstructionService_StreamInstructionServer) error {
	meta := req.GetMetadata()
	result, trace := router.Dispatch(req.GetSessionId(), req.GetInstruction(), meta)
	stream.Send(&proto.InstructionResponse{
		Content:  result,
		Trace:    trace,
		Complete: true,
	})
	return nil
}
