package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type InstanceParams struct {
	Instance string `json:"instance" jsonschema:"The name of the Tempo instance to query"`
	Tenant   string `json:"tenant" jsonschema:"The tenant to query. This field is only required for multi-tenant Tempo instances."`
}

func ptr[T any](v T) *T { return &v }

func withInstanceParams(params map[string]*jsonschema.Schema) map[string]*jsonschema.Schema {
	params["instance"] = &jsonschema.Schema{
		Type:        "string",
		MinLength:   ptr(1),
		Description: "The name of the Tempo instance to query",
	}
	params["tenant"] = &jsonschema.Schema{
		Type:        "string",
		Description: "The tenant to query. This field is only required for multi-tenant Tempo instances.",
	}
	return params
}

func (s *MCPServer) callRemoteTempoMCP(ctx context.Context, instanceParams InstanceParams, toolName string, toolParams any) (*mcp.CallToolResult, error) {
	var instance *InstanceConfig
	for i := range s.instances {
		if s.instances[i].Name == instanceParams.Instance {
			instance = &s.instances[i]
			break
		}
	}

	if instance == nil {
		return nil, fmt.Errorf("Tempo instance '%s' not found", instanceParams.Instance)
	}

	if len(instance.Tenants) > 0 && instanceParams.Tenant == "" {
		return nil, errors.New("tenant name must not be empty")
	}

	endpoint, err := renderEndpointTemplate(instance.Endpoint, instanceParams.Tenant)
	if err != nil {
		return nil, err
	}

	return callRemoteTool(ctx, endpoint, toolName, toolParams)
}

// This needs https://github.com/modelcontextprotocol/go-sdk/pull/634, otherwise it'll block forever
func callRemoteTool(ctx context.Context, url string, toolName string, toolParams any) (*mcp.CallToolResult, error) {
	client := mcp.NewClient(&mcp.Implementation{
		Name:    MCP_NAME,
		Version: MCP_VERSION,
	}, nil)

	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: url}, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	params := &mcp.CallToolParams{
		Name:      toolName,
		Arguments: toolParams,
	}
	return session.CallTool(ctx, params)
}

func renderEndpointTemplate(endpoint string, tenant string) (string, error) {
	templ, err := template.New("endpoint").Parse(endpoint)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = templ.Execute(&buf, map[string]any{
		"Tenant": tenant,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
