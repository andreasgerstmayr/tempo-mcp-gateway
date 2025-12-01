package server

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetTraceParams struct {
	TraceID string `json:"trace_id"`
}

type GetTraceInput struct {
	GetTraceParams
	InstanceParams
}

func (s *MCPServer) RegisterGetTraceTool() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get-trace",
		Description: "Retrieve a specific trace by ID",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: ptr(false),
		},
		InputSchema: &jsonschema.Schema{
			Type:     "object",
			Required: []string{"trace_id", "instance"},
			Properties: withInstanceParams(map[string]*jsonschema.Schema{
				"trace_id": {
					Type:        "string",
					MinLength:   ptr(1),
					Description: "Trace ID to retrieve",
				},
			}),
		},
	}, s.GetTraceTool)
}

func (s *MCPServer) GetTraceTool(ctx context.Context, req *mcp.CallToolRequest, in GetTraceInput) (*mcp.CallToolResult, any, error) {
	result, err := s.callRemoteTempoMCP(ctx, in.InstanceParams, "get-trace", in.GetTraceParams)
	return result, nil, err
}
