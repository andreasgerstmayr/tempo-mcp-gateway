package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/andreasgerstmayr/tempo-mcp-gateway/pkg/server"
	zaplogfmt "github.com/jsternberg/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

func main() {
	config := zap.NewDevelopmentEncoderConfig()
	logger := zap.New(zapcore.NewCore(
		zaplogfmt.NewEncoder(config),
		os.Stdout,
		zapcore.DebugLevel,
	))

	var configPath string
	var listenAddr string
	flag.StringVar(&configPath, "config", "config.yaml", "The path to the Tempo MCP configuration file.")
	flag.StringVar(&listenAddr, "listen", "0.0.0.0:8080", "The listen address of the MCP server.")
	flag.Parse()

	mcpConfig := []server.InstanceConfig{}
	if configPath != "" {
		buff, err := os.ReadFile(configPath)
		if err != nil {
			logger.Fatal("failed to read configFile", zap.String("path", configPath), zap.Error(err))
		}

		err = yaml.UnmarshalStrict(buff, &mcpConfig)
		if err != nil {
			logger.Fatal("failed to parse configFile", zap.String("path", configPath), zap.Error(err))
		}
	}

	logger.Info("Starting Tempo MCP gateway", zap.String("listen", listenAddr))
	server := server.New(logger, mcpConfig)

	err := http.ListenAndServe(listenAddr, server)
	if err != nil {
		logger.Fatal("error", zap.Error(err))
	}
}
