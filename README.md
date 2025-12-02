# tempo-mcp-gateway
A instance-aware and tenant-aware MCP gateway for the [Grafana Tempo MCP server](https://grafana.com/docs/tempo/latest/api_docs/mcp-server/).

## Quickstart
```
kubectl create serviceaccount demo
TOKEN=$(kubectl create token demo)
claude mcp add --transport=http tempo http://tempo-mcp-gateway-openshift-tracing.apps-crc.testing --header "Authorization: Bearer $TOKEN"
```
