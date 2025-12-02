package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/andreasgerstmayr/tempo-mcp-gateway/pkg/mcpserver"
	"github.com/andreasgerstmayr/tempo-mcp-gateway/pkg/tempodiscovery"
	zaplogfmt "github.com/jsternberg/zap-logfmt"
	oscrypto "github.com/openshift/library-go/pkg/crypto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	config := zap.NewDevelopmentEncoderConfig()
	logger := zap.New(zapcore.NewCore(
		zaplogfmt.NewEncoder(config),
		os.Stdout,
		zapcore.DebugLevel,
	))

	var listenAddr string
	var readOnly bool
	flag.StringVar(&listenAddr, "listen", "0.0.0.0:8080", "The listen address of the MCP server.")
	flag.BoolVar(&readOnly, "read-only", false, "Enable this to only expose readonly tools.")
	flag.Parse()

	k8sConfig, errInCluster := rest.InClusterConfig()
	if errInCluster != nil {
		// Try local kubeconfig file
		var errKubeconfig error
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		k8sConfig, errKubeconfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
		if errKubeconfig != nil {
			panic(fmt.Errorf("cannot get in-cluster config: '%w' or kubeconfig: '%w'", errInCluster, errKubeconfig))
		}
	}

	k8sClient, err := client.New(k8sConfig, client.Options{Scheme: tempodiscovery.Scheme})
	if err != nil {
		logger.Fatal("error", zap.Error(err))
	}

	tlsConfig, err := buildTlsConfig("/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt")
	if err != nil {
		logger.Fatal("error", zap.Error(err))
	}

	logger.Info("Starting Tempo MCP gateway", zap.String("listen", listenAddr))
	server := mcpserver.New(logger, k8sClient, tlsConfig, readOnly)

	err = http.ListenAndServe(listenAddr, server.HttpServer)
	if err != nil {
		logger.Fatal("error", zap.Error(err))
	}
}

func buildTlsConfig(serviceCaCertPath string) (*tls.Config, error) {
	serviceCertPEM, err := os.ReadFile(serviceCaCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: tried '%s' and got %v", serviceCaCertPath, err)
	}

	serviceProxyRootCAs := x509.NewCertPool()
	ok := serviceProxyRootCAs.AppendCertsFromPEM(serviceCertPEM)
	if !ok {
		return nil, fmt.Errorf("no CA found for Kubernetes services")
	}

	serviceProxyTLSConfig := oscrypto.SecureTLSConfig(&tls.Config{
		RootCAs: serviceProxyRootCAs,
	})

	return serviceProxyTLSConfig, nil
}
