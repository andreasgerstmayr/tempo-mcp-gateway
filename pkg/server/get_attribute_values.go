package server

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetAttributeValuesParams struct {
	Name        string  `json:"name"`
	FilterQuery *string `json:"filter-query"`
}

type GetAttributeValuesInput struct {
	GetAttributeValuesParams
	InstanceParams
}

func (s *MCPServer) RegisterGetAttributeValuesTool() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get-attribute-values",
		Description: "Get a list of values for a fully scoped attribute name. This is useful for finding the values of a specific attribute. i.e. you can find all the services in the data by asking for resource.service.name",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: ptr(false),
		},
		InputSchema: &jsonschema.Schema{
			Type:     "object",
			Required: []string{"name", "instance"},
			Properties: withInstanceParams(map[string]*jsonschema.Schema{
				"name": {
					Type:        "string",
					MinLength:   ptr(1),
					Description: "The attribute name to get values for (e.g. 'span.http.method', 'resource.service.name')",
				},
				"filter-query": {
					Type:        "string",
					Description: "Filter query to apply to the attribute values. It can only have one spanset and only &&'ed conditions like { <cond> && <cond> && ... }.This is useful for filtering the values to a specific set of values. i.e. you can find all endpoints for a given service by asking for span.http.endpoint and filtering resource.service.name.",
				},
			}),
		},
	}, s.GetAttributeValuesTool)
}

func (s *MCPServer) GetAttributeValuesTool(ctx context.Context, req *mcp.CallToolRequest, in GetAttributeValuesInput) (*mcp.CallToolResult, any, error) {
	result, err := s.callRemoteTempoMCP(ctx, in.InstanceParams, "get-attribute-values", in.GetAttributeValuesParams)
	return result, nil, err
}
