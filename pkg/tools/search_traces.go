package tools

import (
	"context"

	"github.com/andreasgerstmayr/tempo-mcp-server/pkg/tempo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func SearchTraces(tempoClient *tempo.Client) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool(
			"search_traces",
			mcp.WithDescription("Search for traces using a TraceQL formatted query"),
			mcp.WithString("traceql",
				mcp.Required(),
				mcp.Description("search query formatted in TraceQL"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			traceql, err := request.RequireString("traceql")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			trace, err := tempoClient.SearchTraces(ctx, traceql)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(trace), nil
		}
}
