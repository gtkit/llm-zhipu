// @Author xiaozhaofu 2023/8/15 11:02:00
package zhipu

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	DEFULTTIMES = 12
)

type ZpClaims struct {
	APIKey    string `json:"api_key"`
	Exp       int64  `json:"exp"`
	Timestamp int64  `json:"timestamp"`
}

// GenerateToken 生成一个token.
func GenerateToken(apiKey string, duration time.Duration) (string, error) {
	if apiKey == "" {
		return "", errors.New("密钥不能为空")
	}
	if !strings.Contains(apiKey, ".") {
		return "", errors.New("密钥格式不正确")
	}

	apiKeyInfo := strings.Split(apiKey, ".")
	key, secret := apiKeyInfo[0], apiKeyInfo[1]

	if duration == 0 {
		duration = DEFULTTIMES * time.Hour
	}

	return createToken(ZpClaims{
		key,
		time.Now().Add(duration).Unix(),
		time.Now().Unix(),
	}, secret)
}

// createToken 生成一个token.
func createToken(claims ZpClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"api_key":   claims.APIKey,
		"exp":       claims.Exp,
		"timestamp": claims.Timestamp,
	})

	token.Header["alg"] = "HS256"
	token.Header["sign_type"] = "SIGN"
	res, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return res, nil
}
