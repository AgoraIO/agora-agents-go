package option

import (
	core "github.com/AgoraIO/agora-agents-go/v2/core"
)

// WithToken sets the Authorization header to "agora token=<token>" for ConvoAI REST API auth.
func WithToken(token string) *core.AuthorizationOption {
	return &core.AuthorizationOption{
		Authorization: "agora token=" + token,
	}
}

// Area type alias for global regions
type Area = core.Area

// Area constants
const (
	AreaUnknown = core.AreaUnknown
	AreaUS      = core.AreaUS
	AreaEU      = core.AreaEU
	AreaAP      = core.AreaAP
	AreaCN      = core.AreaCN
)

// WithArea creates a new AreaRequestOption with a pool for the specified area.
// The pool manages regional URL cycling and automatic domain selection.
func WithArea(area core.Area) *core.AreaRequestOption {
	return core.NewAreaRequestOption(area)
}

// WithPool creates a new AreaRequestOption with a pre-configured pool.
// Use this when you want to manage the pool lifecycle yourself.
func WithPool(pool *core.Pool) *core.AreaRequestOption {
	return &core.AreaRequestOption{Pool: pool}
}

// WithBasicAuth authenticates REST requests using an Agora Customer ID and Customer Secret.
func WithBasicAuth(customerID, customerSecret string) *core.BasicAuthOption {
	return &core.BasicAuthOption{Username: customerID, Password: customerSecret}
}
