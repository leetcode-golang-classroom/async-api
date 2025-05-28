package util

import (
	"context"

	"github.com/leetcode-golang-classroom/golang-async-api/internal/user"
)

func CloseChannel(ch chan error) {
	if _, ok := <-ch; ok {
		close(ch)
	}
}

type userCtxKey struct{}

func ContextWithUserID(ctx context.Context, user *user.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

func UserFromContext(ctx context.Context) (*user.User, bool) {
	user, ok := ctx.Value(userCtxKey{}).(*user.User)
	if !ok || user == nil {
		return nil, false
	}
	return user, true
}
