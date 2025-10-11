package middlewares

import (
	"github.com/CharafEB/slm-unv/model"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Application struct {
	Srv      *mcp.Server
	Database model.Store
}
