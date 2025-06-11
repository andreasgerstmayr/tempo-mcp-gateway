package tools

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/andreasgerstmayr/tempo-mcp-server/pkg/tempo"
	"github.com/grafana/tempo/pkg/traceql"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const exampleQueries = `
{ name = "GET /api/orders" }
{ resource.service.name = "frontend" }
{ resource.service.name = "frontend" && name = "GET /api/orders" }
{ status = error }
`

func SearchTraces(tempoClient *tempo.Client) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool(
			"searchTraces",
			mcp.WithDescription("Search for traces in Tempo"),
			mcp.WithString("traceql",
				mcp.Required(),
				mcp.Description(fmt.Sprintf("Search query in the TraceQL query language. Example queries:\n%s", exampleQueries)),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := request.RequireString("traceql")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// sanity check the query for syntax errors, and return a list of valid queries if it fails
			_, err = traceql.Parse(query)
			if err != nil {
				slog.Warn("search_traces called with invalid query", "query", query, "error", err)
				return mcp.NewToolResultError(fmt.Errorf("invalid TraceQL query: %w. Please consider this list of valid queries: %s", err, exampleQueries).Error()), nil
			}

			trace, err := tempoClient.SearchTraces(ctx, query)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(trace), nil
		}
}
