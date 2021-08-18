package main

import (
	"github.com/dgrijalva/jwt-go"
)

func ParseToken(tokenString string) (*jwt.Token, *Claims, error) {
	Claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, Claims, func(token *jwt.Token) (i interface{}, err error) {
		return jwtkey, nil
	})
	return token, Claims, err
}
func Getting(tokenString string) *Claims {
	//vcalidate token formate
	if tokenString == "" {
		// ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "权限不足"})
		// ctx.Abort()
		return nil
	}

	token, claims, err := ParseToken(tokenString)
	if err != nil || !token.Valid {
		// ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "权限不足"})
		// ctx.Abort()
		return nil
	}
	// fmt.Println(111)
	// fmt.Println(claims.UserId)
	return claims
}

type Claims struct {
	UserId string
	RoomId string

	jwt.StandardClaims
}

var jwtkey = []byte("hello world")
