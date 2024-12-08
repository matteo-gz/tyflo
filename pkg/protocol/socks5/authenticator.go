package socks5

import (
	"context"
	"errors"
	"fmt"
)

type Authenticator interface {
	// Authenticate 验证用户名和密码
	// 返回nil表示验证通过,否则返回错误
	Authenticate(ctx context.Context, username, password string) error
}

// NoAuthenticator 无需认证的认证器
type NoAuthenticator struct{}

func (NoAuthenticator) Authenticate(ctx context.Context, username, password string) error {
	return nil
}

// UserPassAuthenticator 用户名密码认证器
type UserPassAuthenticator struct {
	users map[string]string
}

// NewUserPassAuthenticator 创建一个新的用户名密码认证器
func NewUserPassAuthenticator(users map[string]string) *UserPassAuthenticator {
	return &UserPassAuthenticator{
		users: users,
	}
}

func (a *UserPassAuthenticator) Authenticate(ctx context.Context, username, password string) error {
	if pass, exists := a.users[username]; exists {
		if pass == password {
			return nil
		}
	}
	return errors.New(fmt.Sprintf("authentication failed for [%s]:[%s]", username, password))
}
