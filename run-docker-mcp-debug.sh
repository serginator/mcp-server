#!/bin/bash

# Debug wrapper script to run MCP server in Docker for Cursor integration
# This allows Cursor to spawn the Docker container as if it were a local process

cd "$(dirname "$0")"

# Log the input for debugging
exec 3>&1
exec 1> >(tee -a /tmp/mcp-debug.log)
exec 2>&1

echo "$(date): Starting MCP Docker container" >&3
echo "$(date): Working directory: $(pwd)" >&3
echo "$(date): Input received:" >&3

# Save input to a temp file and also pass it through
tee /tmp/mcp-input.log | docker run --rm -i \
  -v "$(pwd)/config.yml:/app/config.yml:ro" \
  mcp-test-server:latest

echo "$(date): Docker container finished" >&3


