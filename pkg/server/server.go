package server

import (
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

const MCP_NAME = "tempo-mcp-gateway"
const MCP_VERSION = "v1.0.0"

type MCPServer struct {
	logger    *zap.Logger
	instances []InstanceConfig

	mcpServer   *mcp.Server
	httpHandler *mcp.StreamableHTTPHandler
}

type InstanceConfig struct {
	Name     string   `yaml:"name"`
	Endpoint string   `yaml:"endpoint"`
	Tenants  []string `yaml:"tenants,omitempty"`
}

func New(logger *zap.Logger, instances []InstanceConfig) *MCPServer {
	server := mcp.NewServer(&mcp.Implementation{Name: MCP_NAME, Version: MCP_VERSION}, nil)

	httpHandler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return server
	}, nil)

	s := &MCPServer{
		logger:      logger,
		instances:   instances,
		mcpServer:   server,
		httpHandler: httpHandler,
	}

	s.RegisterListInstancesTool()
	s.RegisterListTenantsTool()
	s.RegisterSearchTool()
	s.RegisterGetTraceTool()
	s.RegisterGetAttributeNamesTool()
	s.RegisterGetAttributeValuesTool()

	return s
}

func (s *MCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.httpHandler.ServeHTTP(w, r)
}
