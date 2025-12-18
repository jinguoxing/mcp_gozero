# Quick Start Guide

**Feature**: MCP Tool for go-zero Framework
**Version**: 1.0.0
**Date**: November 14, 2025

## Overview

mcp-zero is an MCP (Model Context Protocol) server that brings go-zero framework capabilities to AI assistants like Claude Desktop. Use natural language to create services, generate code, and manage go-zero projects.

## Prerequisites

- Go 1.19 or later
- goctl (go-zero CLI tool)
- Claude Desktop (or other MCP-compatible AI assistant)

## Installation

### 1. Install goctl

```bash
go install github.com/zeromicro/go-zero/tools/goctl@latest
```

Verify installation:

```bash
goctl --version
```

### 2. Install mcp-zero

Build from source:

```bash
git clone https://github.com/jinguoxing/mcp-gozero.git
cd mcp-gozero
go build -o mcp-gozero
```

After the project is published, you'll be able to install via:

```bash
go install github.com/jinguoxing/mcp-gozero@latest
```

### 3. Configure Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "go-zero": {
      "command": "/Users/yourname/go/bin/mcp-gozero",
      "env": {
        "GOCTL_PATH": "/Users/yourname/go/bin/goctl"
      }
    }
  }
}
```

**Note**: Adjust paths to match your installation locations.

### 4. Restart Claude Desktop

Quit and restart Claude Desktop to load the new MCP server.

## Quick Tutorial

### Creating Your First API Service

Open Claude Desktop and try:

```text
Create a new API service called "userservice" on port 8080
```

Claude will use mcp-zero to:

1. Validate the service name
2. Execute goctl to generate project structure
3. Fix import paths and initialize go modules
4. Verify the build succeeds
5. Provide next steps

Example response:

```text
Successfully created api service 'userservice'

Output directory: /your/current/directory/userservice

Additional Information:
  port: 8080
  style: go_zero

Next steps:
  1. cd /your/current/directory/userservice
  2. go mod tidy
  3. go run .
```

### Understanding the Generated Structure

The generated service includes:
- `userservice.api` - API specification file
- `etc/userservice.yaml` - Configuration file
- `internal/handler/` - HTTP handlers
- `internal/logic/` - Business logic
- `internal/svc/servicecontext.go` - Service context
- `userservice.go` - Main entry point

### Adding Your First Endpoint

Before running the service, you need to define your API endpoints. Edit the `userservice.api` file:

```go
syntax = "v1"

info (
    title: "userservice"
    desc: "userservice API"
    author: "your name"
    email: "your@email.com"
)

type (
    PingRequest {}
    PingResponse {
        Message string `json:"message"`
    }
)

service userservice-api {
    @handler PingHandler
    get /ping (PingRequest) returns (PingResponse)
}
```

Then regenerate the code:

```text
Generate code from the userservice.api file in ./userservice
```

### Running the Service

```bash
cd userservice
go run userservice.go
```

Now test it:

```bash
curl http://localhost:8080/ping
# Response: {"message":"pong"}
```

You'll need to implement the logic in `internal/logic/pinglogic.go`:

```go
func (l *PingLogic) Ping(req *types.PingRequest) (resp *types.PingResponse, err error) {
    return &types.PingResponse{
        Message: "pong",
    }, nil
}
```

### Adding More Endpoints

To add another endpoint, just add it to the `.api` file and regenerate:

```go
service userservice-api {
    @handler PingHandler
    get /ping (PingRequest) returns (PingResponse)

    @handler UserInfoHandler
    get /user/:id (UserInfoRequest) returns (UserInfoResponse)
}
```

### Creating an RPC Service

```text
Create an RPC service called "authservice" with methods Login and Verify
```

### Analyzing a Project

```text
Analyze the project in /path/to/my/go-zero/project
```

You'll get:

- List of all endpoints
- Service dependencies
- Configuration files
- Suggestions for improvements

### Generating Database Models

```text
Generate models for the users table in mysql://user:pass@localhost:3306/mydb
```

### Creating Middleware

```text
Generate authentication middleware for JWT tokens
```

## Common Use Cases

### 1. Starting a New Microservice Project

```text
Create these services in ./services directory:
- API gateway on port 8080
- User service RPC on port 9001
- Order service RPC on port 9002
```

### 2. Spec-First Development

```text
Create an API spec for a user service with these endpoints:
- POST /login
- POST /register
- GET /user/:id
- PUT /user/:id
```

Then generate code:

```text
Generate code from userservice.api
```

### 3. Adding Features to Existing Service

```text
I have a service at ./userservice. Add rate limiting middleware.
```

### 4. Configuration Management

```text
Generate a production configuration template for my API service
```

### 5. Migration from Another Framework

```text
How do I migrate my Express.js API to go-zero?
```

## Tips & Best Practices

### Service Naming

- ✅ Good: `userservice`, `orderservice`, `apigateway`
- ❌ Bad: `user-service` (hyphens not allowed), `123service` (must start with letter)

### Port Selection

- Use ports 8080-8089 for API services
- Use ports 9000-9099 for RPC services
- mcp-zero will warn if ports are already in use

### Project Organization

For monorepo:

```text
myproject/
├── services/
│   ├── api-gateway/
│   ├── user-service/
│   └── order-service/
└── shared/
    └── types/
```

Ask Claude:

```text
Create services in ./services directory
```

### Error Recovery

If generation fails:

```text
The service generation failed. Try again with corrected parameters.
```

mcp-zero preserves partial state, so you won't lose progress.

### Getting Help

```text
Explain go-zero middleware
```

```text
What's the difference between API and RPC services in go-zero?
```

```text
Show me best practices for error handling in go-zero
```

## Troubleshooting

### "goctl not found"

**Problem**: mcp-zero can't find goctl executable.

**Solution**:

1. Verify goctl is installed: `goctl --version`
2. Add GOCTL_PATH to Claude config (see Installation step 3)
3. Or add goctl to your PATH

### "Service name validation failed"

**Problem**: Service name contains invalid characters.

**Solution**: Use only letters, numbers, and underscores. Start with a letter.

- `user-service` → `userservice` or `user_service`

### "Port already in use"

**Problem**: Specified port is occupied.

**Solution**: Choose a different port or stop the service using that port.

### "Build failed"

**Problem**: Generated code doesn't compile.

**Solution**:

1. Check the error message for missing dependencies
2. Run `go mod tidy`
3. Report the issue if it persists

### "Permission denied"

**Problem**: Can't write to output directory.

**Solution**: Use a directory where you have write permissions, or create the directory first.

## Advanced Usage

### Custom Templates

```text
Use this middleware template for my service:
[paste your template]
```

### Batch Operations

```text
Create 5 microservices: user, order, product, payment, notification
All RPC services, ports starting from 9001
```

### Configuration Validation

```text
Validate my configuration file at ./etc/config.yaml
```

### Project Documentation

```text
Generate documentation for my project structure
```

## Environment Variables

- `GOCTL_PATH`: Override goctl executable location

**Happy coding with mcp-gozero! 🚀**
