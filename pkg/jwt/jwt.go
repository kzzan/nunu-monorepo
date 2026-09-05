// Package jwt 提供 JWT 签发与解析能力。
package jwt

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/do/v2"
	"github.com/spf13/viper"
)

// JWT 是 HS256 签名的 JWT 工具，密钥来自配置 security.jwt.key。
type JWT struct {
	key []byte
}

// MyCustomClaims 是业务自定义声明：携带用户 ID。
type MyCustomClaims struct {
	UserId uint
	jwt.RegisteredClaims
}

// Package registers the jwt provider into the injector.
var Package = do.Package(do.Lazy(New))

// New 从配置构造 JWT 工具，由注入容器调用。
func New(i do.Injector) (*JWT, error) {
	conf := do.MustInvoke[*viper.Viper](i)
	return &JWT{key: []byte(conf.GetString("security.jwt.key"))}, nil
}

// GenToken 为指定用户签发 JWT，过期时间由调用方决定。
func (j *JWT) GenToken(userId uint, expiresAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, MyCustomClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "",
			Subject:   "",
			ID:        "",
			Audience:  []string{},
		},
	})

	// Sign and get the complete encoded token as a string using the key
	tokenString, err := token.SignedString(j.key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ParseToken 解析并校验 JWT（自动剥除 Bearer 前缀），
// 返回声明或校验错误。
func (j *JWT) ParseToken(tokenString string) (*MyCustomClaims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	if strings.TrimSpace(tokenString) == "" {
		return nil, errors.New("token is empty")
	}
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.key, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}
