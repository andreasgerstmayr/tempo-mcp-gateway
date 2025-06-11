package tools

import (
	"context"

	"github.com/andreasgerstmayr/tempo-mcp-server/pkg/tempo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func ShowStatus(tempoClient *tempo.Client) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool(
			"show_status",
			mcp.WithDescription("Show the status of all Tempo services"),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			status, err := tempoClient.StatusServices(ctx)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(status), nil
		}
}
