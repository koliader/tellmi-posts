package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/koliader/tellmi-posts/internal/lib/token"
	"google.golang.org/grpc/metadata"
)

const admin = "ADMIN"

func (m *Middleware) AuthorizeAdmin(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}
	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, fmt.Errorf("missing auth header")
	}
	authHeader := values[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid auth header format")
	}
	tokenString := fields[1]
	payload, err := m.tokenMaker.VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}
	if payload.Role != admin {
		return nil, fmt.Errorf("no access")
	}
	return payload, nil
}
