package server

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetAttributeNamesParams struct {
	Scope *string `json:"scope"`
}

type GetAttributeNamesInput struct {
	GetAttributeNamesParams
	InstanceParams
}

func (s *MCPServer) RegisterGetAttributeNamesTool() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get-attribute-names",
		Description: "Get a list of available attribute names that can be used in TraceQL queries. This is useful for finding the names of attributes that can be used in a query.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: ptr(false),
		},
		InputSchema: &jsonschema.Schema{
			Type:     "object",
			Required: []string{"instance"},
			Properties: withInstanceParams(map[string]*jsonschema.Schema{
				"scope": {
					Type:        "string",
					Description: "Optional scope to filter attributes by (span, resource, event, link, instrumentation). If not provided, returns all attributes.",
				},
			}),
		},
	}, s.GetAttributeNamesTool)
}

func (s *MCPServer) GetAttributeNamesTool(ctx context.Context, req *mcp.CallToolRequest, in GetAttributeNamesInput) (*mcp.CallToolResult, any, error) {
	result, err := s.callRemoteTempoMCP(ctx, in.InstanceParams, "get-attribute-names", in.GetAttributeNamesParams)
	return result, nil, err
}
