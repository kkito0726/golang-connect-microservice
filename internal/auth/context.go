package auth

import "context"

// contextKey は他パッケージのキーと衝突しないよう非公開型を使う。
type contextKey struct{ name string }

var (
	userIDKey = &contextKey{"userID"}
	roleKey   = &contextKey{"role"}
	tokenKey  = &contextKey{"token"} // サービス間転送用の生トークン
)

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok
}

func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey, role)
}

func RoleFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(roleKey).(string)
	return v, ok
}

// WithToken は検証済みの生トークン文字列を ctx に保存する。
// サービス間通信でトークンを転送する際に使用する。
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

// TokenFromContext は ctx から生トークン文字列を取り出す。
func TokenFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(tokenKey).(string)
	return v, ok
}
