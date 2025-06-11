package main

import (
	"flag"
	"log/slog"

	"github.com/andreasgerstmayr/tempo-mcp-server/pkg/resources"
	"github.com/andreasgerstmayr/tempo-mcp-server/pkg/tempo"
	"github.com/andreasgerstmayr/tempo-mcp-server/pkg/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	var tempoApiUrl string
	flag.StringVar(&tempoApiUrl, "tempo-api-url", "http://localhost:3200", "URL of Tempo API")
	flag.Parse()

	slog.Info("Starting Tempo MCP server", "tempoApiUrl", tempoApiUrl)
	mcpServer := server.NewMCPServer("Tempo MCP server", "1.0.0")
	tempoClient := tempo.NewClient(tempoApiUrl)

	// Note: MCP Resources are not supported in VS Code
	mcpServer.AddResource(resources.QueryTraceById(tempoClient))

	mcpServer.AddTool(tools.QueryTraceById(tempoClient))
	mcpServer.AddTool(tools.SearchTraces(tempoClient))
	mcpServer.AddTool(tools.ShowStatus(tempoClient))

	err := server.ServeStdio(mcpServer)
	if err != nil {
		slog.Error("Server error", "error", err)
	}
}
