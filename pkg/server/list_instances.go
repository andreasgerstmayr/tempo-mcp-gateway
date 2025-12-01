package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListInstancesOutput struct {
	Instances []string `json:"instances"`
}

func (s *MCPServer) RegisterListInstancesTool() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list-instances",
		Description: "List all Tempo instances",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: ptr(false),
		},
	}, s.handleListInstances)
}

func (s *MCPServer) handleListInstances(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, ListInstancesOutput, error) {
	// TODO: query Kubernetes API for TempoStack and TempoMonolithic instances
	instances := make([]string, len(s.instances))
	for i := range s.instances {
		instances[i] = s.instances[i].Name
	}

	return nil, ListInstancesOutput{Instances: instances}, nil
}
