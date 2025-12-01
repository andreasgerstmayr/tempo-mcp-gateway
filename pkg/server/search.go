package server

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SearchParams struct {
	Query string  `json:"query"`
	Start *string `json:"start"`
	End   *string `json:"end"`
}

type SearchInput struct {
	SearchParams
	InstanceParams
}

func (s *MCPServer) RegisterSearchTool() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "traceql-search",
		Description: "Search for traces using TraceQL queries",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: ptr(false),
		},
		InputSchema: &jsonschema.Schema{
			Type:     "object",
			Required: []string{"query", "instance"},
			Properties: withInstanceParams(map[string]*jsonschema.Schema{
				"query": {
					Type:        "string",
					MinLength:   ptr(1),
					Description: "TraceQL query string",
				},
				"start": {
					Type:        "string",
					Description: "Start time for the search (RFC3339 format). If not provided will search the past 1 hour. If provided, must be before end.",
				},
				"end": {
					Type:        "string",
					Description: "End time for the search (RFC3339 format). If not provided will search the past 1 hour. If provided, must be after start.",
				},
			}),
		},
	}, s.SearchTool)
}

func (s *MCPServer) SearchTool(ctx context.Context, req *mcp.CallToolRequest, in SearchInput) (*mcp.CallToolResult, any, error) {
	result, err := s.callRemoteTempoMCP(ctx, in.InstanceParams, "traceql-search", in.SearchParams)
	return result, nil, err
}
