package main

import (
	"context"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	UserId string
	RoomId string

	jwt.StandardClaims
}

var jwtkey = []byte("hello world")
var ctx = context.Background()

// func Client() (string, uint32) {
// 	conn, err := grpc.Dial(":1234", grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	client := NewRCenterClient(conn)
// 	roomlist, err := client.GetRoomList(ctx, &ReqGetRoomList{})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	roomids := roomlist.GetRoomIds()
// 	roomid, err := client.GetRoom(ctx, &ReqGetRoom{RoomId: roomids[0], Num: 1})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return roomid.GetAddr(), roomid.GetRoomId()
// 	//roomid.GetRoomId()
// 	//playernums := roomlist.GetNums()
// }
func GetToken(userid string, roomid string) string {
	//fmt.Println(userid)
	expireTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims{
		UserId: userid,
		RoomId: roomid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(), //过期时间
			IssuedAt:  time.Now().Unix(),
			Issuer:    "127.0.0.1",  // 签名颁发者
			Subject:   "user token", //签名主题
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, err := token.SignedString(jwtkey)
	if err != nil {
		log.Fatal(err)
	}
	return tokenstring
}
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
