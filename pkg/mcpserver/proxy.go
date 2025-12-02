package mcpserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

type contextKey string

const authTokenKey contextKey = "authToken"

func WithAuthTokenFromHeader(ctx context.Context, header http.Header) context.Context {
	rawToken := header.Get("Authorization")
	token := strings.TrimPrefix(rawToken, "Bearer ")
	return context.WithValue(ctx, authTokenKey, token)
}

func AuthTokenFromContext(ctx context.Context) string {
	value := ctx.Value(authTokenKey)
	token := value.(string)
	return token
}

func (s *MCPServer) listRemoteTools(ctx context.Context, endpoint string) (*mcp.ListToolsResult, error) {
	mcpClient, err := s.createMcpClient(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer mcpClient.Close()

	toolsRequest := mcp.ListToolsRequest{}
	toolsResult, err := mcpClient.ListTools(ctx, toolsRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools from remote MCP server: %w", err)
	}

	return toolsResult, nil
}

func (s *MCPServer) callRemoteTool(ctx context.Context, endpoint string, toolName string, args map[string]any) (*mcp.CallToolResult, error) {
	mcpClient, err := s.createMcpClient(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer mcpClient.Close()

	// Remove additional arguments which are not present in downstream MCP server
	forwardArgs := make(map[string]any)
	for k, v := range args {
		if k != "instance" && k != "tenant" {
			forwardArgs[k] = v
		}
	}

	toolRequest := mcp.CallToolRequest{}
	toolRequest.Params.Name = toolName
	toolRequest.Params.Arguments = forwardArgs
	result, err := mcpClient.CallTool(ctx, toolRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools from remote MCP server: %w", err)
	}

	return result, nil
}

func (s *MCPServer) createMcpClient(ctx context.Context, endpoint string) (*client.Client, error) {
	headers := map[string]string{}
	authToken := AuthTokenFromContext(ctx)
	if authToken != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", authToken)
	}

	httpTransport, err := transport.NewStreamableHTTP(endpoint,
		transport.WithHTTPHeaders(headers),
		transport.WithHTTPBasicClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: s.tlsConfig,
			},
		}),
	)
	mcpClient := client.NewClient(httpTransport)

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    MCP_NAME,
		Version: MCP_VERSION,
	}

	_, err = mcpClient.Initialize(ctx, initReq)
	if err != nil {
		_ = mcpClient.Close()
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	return mcpClient, nil
}
