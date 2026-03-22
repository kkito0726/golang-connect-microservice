package middleware

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	"github.com/ken/connect-microservice/internal/auth"
)

// NewAuthInterceptor は JWT 認証を行う ConnectRPC Interceptor を返す。
// publicProcedures に含まれるプロシージャはトークンなしで通過する。
//
// 使用例:
//
//	connect.WithInterceptors(middleware.NewAuthInterceptor(tokenGen, []string{
//	    "/user.v1.UserService/Login",
//	    "/user.v1.UserService/CreateUser",
//	}))
func NewAuthInterceptor(tokenGen *auth.TokenGenerator, publicProcedures []string) connect.UnaryInterceptorFunc {
	skip := make(map[string]bool, len(publicProcedures))
	for _, p := range publicProcedures {
		skip[p] = true
	}

	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if skip[req.Spec().Procedure] {
				return next(ctx, req)
			}

			authHeader := req.Header().Get("Authorization")
			if authHeader == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authorization header required"))
			}

			tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok || tokenString == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid authorization header format, expected: Bearer <token>"))
			}

			claims, err := tokenGen.ValidateToken(tokenString)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid token: %w", err))
			}

			// 後続の handler/usecase で参照できるよう ctx に埋め込む
			ctx = auth.WithUserID(ctx, claims.UserID)
			ctx = auth.WithRole(ctx, claims.Role)
			// サービス間通信でトークンを転送できるよう生トークンも保存
			ctx = auth.WithToken(ctx, tokenString)

			return next(ctx, req)
		}
	}
}
