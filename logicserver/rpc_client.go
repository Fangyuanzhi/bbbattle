package main

import (
	"log"
	"net/rpc"
)

var Req ResPlayer

type Room struct {
	RoomId string `json:"roomid"`
	Token  string `json:"token"`
	Host   string `json:"host"`
	Port   string `json:"port"`
}

func RpcClient(userid string, name string) (string, string, string, error) {
	conn, err := rpc.Dial("tcp", ":1234")
	if err != nil {
		log.Fatal("error")
		return "", "", "", err
	}
	defer conn.Close()
	err = conn.Call("Logic.GetRoomInfo", &ReqPlayer{UserId: userid, Name: name}, &Req)
	if err != nil {
		log.Println(err)
		return "", "", "", err
	}
	return Req.GetRoomId(), Req.GetToken(), Req.GetAddr(), nil
}

// func RpcClient(userid string, name string) (string, string, string, error) {
// 	conn, err := rpc.Dial("tcp", ":1234")
// 	if err != nil {
// 		log.Fatal("error")
// 		return "", "", "", err
// 	}
// 	req := &UserInfo{}
// 	req.UserId = userid
// 	req.Name = name
// 	reqdata, err := json.Marshal(req)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	var retdata string
// 	ret := &MatchInfo{}
// 	err = conn.Call("Match.Start", string(reqdata[:]), &retdata)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	err = json.Unmarshal([]byte(retdata), ret)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	if err != nil {
// 		log.Println(err)
// 		return "", "", "", err
// 	}
// 	return ret.RoomId, ret.Token, ret.Addr, nil
// }
