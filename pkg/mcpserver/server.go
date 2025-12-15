package mcpserver

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/andreasgerstmayr/tempo-mcp-gateway/pkg/tempodiscovery"
	tempov1alpha1 "github.com/grafana/tempo-operator/api/tempo/v1alpha1"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const MCP_NAME = "tempo-mcp-gateway"
const MCP_VERSION = "v1.0.0"

type MCPServer struct {
	logger    *zap.Logger
	k8sClient client.Client
	tlsConfig *tls.Config
	readOnly  bool

	mcpServer        *server.MCPServer
	HttpServer       *server.StreamableHTTPServer
	toolsInitialized bool
}

func New(logger *zap.Logger, k8sClient client.Client, tlsConfig *tls.Config, readOnly bool) *MCPServer {
	hooks := &server.Hooks{}
	mcpServer := server.NewMCPServer(MCP_NAME, MCP_VERSION,
		server.WithToolCapabilities(true),
		server.WithHooks(hooks),
		server.WithInstructions(`
This server provides access to Tempo instances in a Kubernetes cluster.

Do not query across multiple instances unless specifically asked by the user.
Ask the user which Tempo instance to query if the user did not specify it explicitly.
`),
	)
	httpServer := server.NewStreamableHTTPServer(mcpServer, server.WithStateful(false))

	s := &MCPServer{
		logger:    logger,
		k8sClient: k8sClient,
		tlsConfig: tlsConfig,
		readOnly:  readOnly,

		mcpServer:  mcpServer,
		HttpServer: httpServer,
	}

	s.registerTools()
	hooks.OnBeforeListTools = []server.OnBeforeListToolsFunc{func(ctx context.Context, id any, request *mcp.ListToolsRequest) {
		if !s.toolsInitialized {
			ctx := WithAuthTokenFromHeader(ctx, request.Header)
			err := s.registerProxiedTools(ctx)
			if err != nil {
				logger.Error("error listing tools from remote MCP server", zap.Error(err))
				return
			}
			s.toolsInitialized = true
		}
	}}

	// In case the MCP client does not list tools first
	hooks.OnBeforeCallTool = []server.OnBeforeCallToolFunc{func(ctx context.Context, id any, request *mcp.CallToolRequest) {
		if !s.toolsInitialized {
			ctx := WithAuthTokenFromHeader(ctx, request.Header)
			err := s.registerProxiedTools(ctx)
			if err != nil {
				logger.Error("error listing tools from remote MCP server", zap.Error(err))
				return
			}
			s.toolsInitialized = true
		}
	}}

	return s
}

func (s *MCPServer) registerTools() {
	s.mcpServer.AddTool(mcp.NewTool("list-instances",
		mcp.WithDescription("List all Tempo instances. The assistant should display the instances in a table."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx = WithAuthTokenFromHeader(ctx, request.Header)
		instances, err := s.listTempoInstances(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultStructuredOnly(map[string]any{
			"instances": instances,
		}), nil
	})
}

func (s *MCPServer) registerProxiedTools(ctx context.Context) error {
	instances, err := s.listTempoInstances(ctx)
	if err != nil {
		return err
	}

	// Read MCP tools from the first accessible Tempo instance in the cluster
	firstInstance, err := findReadyInstance(instances)
	if err != nil {
		return fmt.Errorf("cannot read tools from Tempo MCP server: %w", err)
	}

	firstTenant := ""
	if len(firstInstance.Tenants) > 0 {
		firstTenant = firstInstance.Tenants[0]
	}

	endpoint := firstInstance.GetMCPEndpoint(firstTenant)
	toolsResp, err := s.listRemoteTools(ctx, endpoint)
	if err != nil {
		return err
	}

	for _, tool := range toolsResp.Tools {
		s.registerProxiedTool(tool)
	}
	return nil
}

func (s *MCPServer) registerProxiedTool(tool mcp.Tool) {
	if s.readOnly && (tool.Annotations.ReadOnlyHint == nil || !*tool.Annotations.ReadOnlyHint) {
		return
	}

	// Add parameters to identify a Tempo instance and tenant
	additionalParameters := []mcp.ToolOption{
		mcp.WithString("tempoNamespace",
			mcp.Required(),
			mcp.Description("The namespace of the Tempo instance to query"),
		),
		mcp.WithString("tempoName",
			mcp.Required(),
			mcp.Description("The name of the Tempo instance to query"),
		),
		mcp.WithString("tenant",
			mcp.Description("The tenant to query. This field is only required for multi-tenant Tempo instances."),
		),
	}
	for _, opt := range additionalParameters {
		opt(&tool)
	}

	s.mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx = WithAuthTokenFromHeader(ctx, request.Header)

		tempoNamespace, err := request.RequireString("tempoNamespace")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if tempoNamespace == "" {
			return mcp.NewToolResultError("tempoNamespace parameter must not be empty"), nil
		}

		tempoName, err := request.RequireString("tempoName")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if tempoName == "" {
			return mcp.NewToolResultError("tempoName parameter must not be empty"), nil
		}

		instances, err := s.listTempoInstances(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		instance, err := findInstanceByName(instances, tempoNamespace, tempoName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if !instance.MCPEnabled {
			var specField string
			switch instance.Kind {
			case tempodiscovery.KindTempoStack:
				specField = ".spec.template.queryFrontend.mcpServer.enabled"
			case tempodiscovery.KindTempoMonolithic:
				specField = ".spec.query.mcpServer.enabled"
			}

			msg := fmt.Sprintf("the MCP server is disabled for this instance. To enable it, set the field %s to true in the %s/%s %s instance",
				specField, instance.Namespace, instance.Name, instance.Kind)
			return mcp.NewToolResultError(msg), nil
		}

		var tenantName string
		if instance.Multitenancy {
			tenantName, err = request.RequireString("tenant")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if tenantName == "" {
				return mcp.NewToolResultError("tenant parameter must not be empty"), nil
			}
		}

		endpoint := instance.GetMCPEndpoint(tenantName)
		return s.callRemoteTool(ctx, endpoint, request.Params.Name, request.GetArguments())
	})
}

func (s *MCPServer) listTempoInstances(ctx context.Context) ([]tempodiscovery.TempoInstance, error) {
	auth := tempodiscovery.Authentication{}
	authToken := AuthTokenFromContext(ctx)
	if authToken != "" {
		auth.BearerToken = authToken
	}

	var verbs []string
	if s.readOnly {
		verbs = []string{"get"}
	} else {
		verbs = []string{"create", "get"}
	}

	return tempodiscovery.New(s.logger, s.k8sClient, s.tlsConfig).ListInstances(ctx, auth, verbs)
}

func findInstanceByName(instances []tempodiscovery.TempoInstance, namespace string, name string) (tempodiscovery.TempoInstance, error) {
	for _, instance := range instances {
		if instance.Namespace == namespace && instance.Name == name {
			return instance, nil
		}
	}

	return tempodiscovery.TempoInstance{}, fmt.Errorf("instance '%s' in namespace '%s' not found", name, namespace)
}

func findReadyInstance(instances []tempodiscovery.TempoInstance) (tempodiscovery.TempoInstance, error) {
	readyStatus := string(tempov1alpha1.ConditionReady)

	for _, instance := range instances {
		if instance.Status == readyStatus {
			return instance, nil
		}
	}

	return tempodiscovery.TempoInstance{}, fmt.Errorf("no Tempo instance is in %s state", readyStatus)
}
