# tempo-mcp-server
A Proof of Concept for a MCP (Model Context Protocol) server for Grafana Tempo.

## Supported tools
* **Query Trace by ID** - Query a specific trace using its trace ID
* **Search Traces** - Search for traces using TraceQL formatted queries
* **Show Status** - Display the current status of all Tempo services 

## Example Prompts
```
search for traces using this query: { status = error }
```

```
show me the trace e4280736b13ed0b2098fdd9f5f465e0b
```

```
show me the status of tempo
```
