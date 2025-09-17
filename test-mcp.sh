#!/bin/bash

# Test script for MCP server
echo "Testing MCP Integration Server"
echo "==============================="

# Test initialize
echo "1. Testing initialize..."
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | ./mcp-server | jq '.result.serverInfo'

echo -e "\n2. Testing tools/list..."
echo -e '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}\n{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./mcp-server | tail -1 | jq '.result.tools | length'

echo -e "\n3. Testing GitHub search repositories..."
echo -e '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}\n{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"github_search_repositories","arguments":{"query":"mcp server"}}}' | timeout 10s ./mcp-server | tail -1 | jq '.result.content[0].text' | head -5

echo -e "\nMCP Server tests completed!"
