package tools

import (
	"context"

	"github.com/andreasgerstmayr/tempo-mcp-server/pkg/tempo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func QueryTraceById(tempoClient *tempo.Client) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool(
			"queryTraceByTraceId",
			mcp.WithDescription("Query a trace by trace ID"),
			mcp.WithString("traceid",
				mcp.Required(),
				mcp.Description("TraceID of the trace to query"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			traceid, err := request.RequireString("traceid")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			trace, err := tempoClient.QueryTraceByID(ctx, traceid)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(trace), nil
		}
}
