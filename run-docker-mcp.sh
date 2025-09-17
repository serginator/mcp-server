#!/bin/bash

# Wrapper script to run MCP server in Docker for Cursor integration
# This allows Cursor to spawn the Docker container as if it were a local process

cd "$(dirname "$0")"
docker run --rm -i \
  -v "$(pwd)/config.yml:/app/config.yml:ro" \
  mcp-test-server:latest
