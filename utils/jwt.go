package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MyClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var MySecret = []byte("secret")

// 生成Token
func GenToken(ID uint, name string) (string, error) {
	claims := MyClaims{
		UserID:   ID,
		Username: name,
		RegisteredClaims: jwt.RegisteredClaims{
			//过期时间 24小时
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			//签发人
			Issuer: "myserver",
		},
	}
	// alg:HS256 payload:claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	// 返回签名后的token字符串
	return token.SignedString(MySecret)
}

// 解析token
func ParseToken(tokenString string) (*MyClaims, error) {
	// 解析token 验证签名是否合法 将Payload解码并写入MyClaims结构体
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法 签名算法为HMAC类时返回MySecret
		// 防止黑客篡改token的alg为none算法 直接通过了验证
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的token")
		}
		return MySecret, nil
	})
	if err != nil {
		return nil, err
	}
	// token的Claims字段断言成*MyClaims claims接收 并验证token是否有效
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("无效的token")
}
