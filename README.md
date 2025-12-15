# tempo-mcp-gateway
A instance-aware and tenant-aware MCP gateway for the [Grafana Tempo MCP server](https://grafana.com/docs/tempo/latest/api_docs/mcp-server/).

## Quickstart
```
kubectl create serviceaccount demo
TOKEN=$(kubectl create token demo)
echo "Bearer $TOKEN"
claude mcp remove tempo
claude mcp add --transport=http tempo http://tempo-mcp-gateway-openshift-tracing.apps-crc.testing --header "Authorization: Bearer $TOKEN"
```

## Acknowledgements
* https://github.com/grafana/mcp-grafana
* https://github.com/grafana/tempo
