package api

import (
	"context"

	"github.com/Gradient-Linux/concave/internal/auth"
	"github.com/Gradient-Linux/concave/internal/system"
)

var (
	authenticateUser = auth.Authenticate
	refreshSession   = auth.RefreshToken
	validateSession  = auth.ValidateToken
	runPrivileged    = system.RunPrivileged
	resolveRole      = auth.ResolveRole
)

func resetAPIDeps() {
	authenticateUser = auth.Authenticate
	refreshSession = auth.RefreshToken
	validateSession = auth.ValidateToken
	runPrivileged = system.RunPrivileged
	resolveRole = auth.ResolveRole
}

func withTestClaims(claims auth.Claims, fn func(context.Context)) {
	fn(context.WithValue(context.Background(), claimsContextKey{}, claims))
}
