package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/rpc"
	"time"

	redis "github.com/go-redis/redis/v8"
)

var op = redis.ZRangeBy{
	Min:    "1",
	Max:    "10",
	Offset: 0,
	Count:  1,
}

type Logic struct{}
type Match struct{}
type UserInfo struct {
	UserId string `json:"userid"`
	Name   string `json:"name"`
}
type MatchInfo struct {
	RoomId string `json:"roomid"`
	Addr   string `json:"addr"`
	Token  string `json:"token"`
}

func (s *Match) Start(request string, reply *string) error {
	values, err := rdb.ZRangeByScore(ctx, "room", &op).Result()
	if err != nil {
		log.Println("匹配失败")
		return err
	}
	data := &UserInfo{}
	err = json.Unmarshal([]byte(request), data)
	if err != nil {
		log.Println("unmashal failed", err)
		panic(err)
	}
	//fmt.Println(values[0])
	rdb.ZIncr(ctx, "room", &redis.Z{Score: float64(-1), Member: values[0]}).Result()
	roomid := values[0]
	userid := data.UserId
	token := GetToken(userid, roomid)
	_, err = rdb.Set(ctx, roomid+":token", token, time.Hour*1).Result()
	if err != nil {
		log.Println(err)
		return err
	}
	addr, err := rdb.Get(ctx, roomid+":addr").Result()
	if err != nil {
		log.Println(err)
		return err
	}
	rdata := &MatchInfo{}
	rdata.Addr = addr
	rdata.RoomId = roomid
	rdata.Token = token
	rbdata, err := json.Marshal(rdata)
	if err != nil {
		log.Fatal(err)
	}
	*reply = string(rbdata[:])
	return nil
}

func RpcHandle() {
	err := rpc.RegisterName("Logic", new(Logic))
	if err != nil {
		log.Println(1, err)
		return
	}
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Println(2, err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
		}
		go rpc.ServeConn(conn)
	}
}

func (l *Logic) GetRoomInfo(request *ReqPlayer, reply *ResPlayer) error {
	//fmt.Println("111")
	ok, err := rdb.HExists(ctx, "token", request.UserId).Result()
	if err != nil {
		log.Println(err)
		return err
	}
	if ok {
		token := redismgr.GetToken(request.UserId)
		claims := Getting(token)
		roomid := claims.RoomId
		addr, err := rdb.HGet(ctx, "room:addr", roomid).Result()
		if err != nil {
			log.Println(err)
			return err
		}
		reply.Addr = addr
		reply.Token = token
		reply.RoomId = roomid
		return nil
	}
	values, err := rdb.ZRangeByScore(context.Background(), "room", &op).Result()
	if err != nil {
		log.Println("匹配失败")
		return err
	}
	//fmt.Println(values[0])
	rdb.ZIncr(ctx, "room", &redis.Z{Score: float64(-1), Member: values[0]}).Result()
	roomid := values[0]
	userid := request.GetUserId()
	token := GetToken(userid, roomid)
	//fmt.Println(token)
	ok = redismgr.SetToken(request.Name, token)
	//_, err = rdb.HSet(ctx, "token", request.Name, token).Result()
	if !ok {
		log.Println("token设置失败")
		return nil
	}
	addr, err := rdb.HGet(ctx, "room:addr", roomid).Result()
	if err != nil {
		log.Println(err)
		return err
	}
	reply.Addr = addr
	reply.RoomId = roomid
	reply.Token = token
	return nil
}
