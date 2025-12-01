package server

import (
	"context"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListTenantsInput struct {
	Instance string `json:"instance"`
}

type ListTenantsOutput struct {
	Tenants []string `json:"tenants"`
}

func (s *MCPServer) RegisterListTenantsTool() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list-tenants",
		Description: "List all tenants for a given Tempo instance",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: ptr(false),
		},
		InputSchema: &jsonschema.Schema{
			Type:     "object",
			Required: []string{"instance"},
			Properties: map[string]*jsonschema.Schema{
				"instance": {
					Type:        "string",
					MinLength:   ptr(1),
					Description: "The name of the Tempo instance",
				},
			},
		},
	}, s.handleListTenants)
}

func (s *MCPServer) handleListTenants(ctx context.Context, req *mcp.CallToolRequest, in ListTenantsInput) (*mcp.CallToolResult, ListTenantsOutput, error) {
	// configSchema := &jsonschema.Schema{
	// 	Type: "object",
	// 	Properties: map[string]*jsonschema.Schema{
	// 		"instance": {Type: "string", Description: "TempoStack or TempoMonolithic name"},
	// 	},
	// 	Required: []string{"instance"},
	// }

	// result, err := req.Session.Elicit(ctx, &mcp.ElicitParams{
	// 	Message:         "Please specify an instance",
	// 	RequestedSchema: configSchema,
	// })
	// if err != nil {
	// 	return nil, ListTenantsOutput{}, err
	// }
	// fmt.Printf("result is %v\n", result.Content["instance"])

	// TODO: query Kubernetes API to get the tenants of the TempoStack or TempoMonolithic instance
	var instance *InstanceConfig
	for i := range s.instances {
		if s.instances[i].Name == in.Instance {
			instance = &s.instances[i]
			break
		}
	}

	if instance == nil {
		return nil, ListTenantsOutput{}, fmt.Errorf("instance '%s' not found", in.Instance)
	}

	if len(instance.Tenants) == 0 {
		return nil, ListTenantsOutput{}, fmt.Errorf("instance '%s' is a single-tenant Tempo instance", in.Instance)
	}

	return nil, ListTenantsOutput{Tenants: instance.Tenants}, nil
}
