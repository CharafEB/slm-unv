package middlewares

import (
	"github.com/CharafEB/slm-unv/model"
	"github.com/mark3labs/mcphost/sdk"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/redis/go-redis/v9"
)

type Application struct {
	Srv      *mcp.Server
	Database model.Store
}

type Mcphost struct {
	Host    *sdk.MCPHost
	Redis   *redis.Client
	Channel string
}
