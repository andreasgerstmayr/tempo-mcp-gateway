package resources

import (
	"context"
	"errors"
	"strings"

	"github.com/andreasgerstmayr/tempo-mcp-server/pkg/tempo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func QueryTraceById(tempoClient *tempo.Client) (mcp.Resource, server.ResourceHandlerFunc) {
	return mcp.NewResource(
			"traces://{trace_id}",
			"Trace",
			mcp.WithResourceDescription("Distributed trace information"),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			parts := strings.Split(request.Params.URI, "://")
			if len(parts) != 2 {
				return nil, errors.New("invalid URI format. Expected 'traces://{trace_id}'")
			}
			traceid := parts[1]

			trace, err := tempoClient.QueryTraceByID(ctx, traceid)
			if err != nil {
				return nil, err
			}

			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "application/json",
					Text:     trace,
				},
			}, nil
		}
}
