package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"google.golang.org/grpc"
)

var liveroom = make(map[string]int32)
var rcclient RCenterClient

func RCenterClientInit() {
	conn, err := grpc.Dial("localhost:1235", grpc.WithInsecure())
	if err != nil {
		log.Println("grpc error", err)
		return
	}
	//defer conn.Close()
	rcclient = NewRCenterClient(conn)
}

//rcent异常宕机后同步room信息
func RctUpdateRoomList() {
	for {
		stream, err := rcclient.UpdateRoomList(ctx)
		if err != nil {
			log.Println(err)
			continue
		}
		var (
			roomids []string
			nums    []int32
		)
		for key, value := range liveroom {
			roomids = append(roomids, key)
			nums = append(nums, value)
		}
		stream.Send(&ReqUpdateRoomList{
			RoomIds: roomids,
			Nums:    nums,
		})
	}
}
func RctGettoken(name string) (*RetGetToken, error) {
	// conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	// if err != nil {
	// 	log.Println("grpc error", err)
	// 	return nil, err
	// }
	// defer conn.Close()
	//client := NewRCenterClient(conn)
	token_stream, err := rcclient.GetToken(context.Background())
	if err != nil {
		log.Println("token_stream error", err)
		return nil, err
	}
	err = token_stream.Send(&ReqGetToken{Name: name})
	if err != nil {
		log.Println(err)
	}
	token, err := token_stream.Recv()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return token, nil
}

//一局游戏结束把房间资源释放出来
func RctDelRoom() {
	stream, err := rcclient.DelRoom(context.Background())
	if err != nil {
		log.Println(err)
		return
	}
	//这里后面加个chan便于并发
	for {
		roomid := <-Chan_DelRoom
		fmt.Printf("roomid: %v\n", roomid)
		err := stream.Send(&ReqDelRoom{RoomId: roomid})
		if err != nil {
			continue
		}
	}
	//stream.Send(&ReqDelRoom{RoomId: roomid})
}

//这个函数暂时不用
func RctGetRoom(roomid string) {
	stream, err := rcclient.GetRoom(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	stream.Send(&ReqGetRoom{RoomId: roomid})
}

//room还未开始游戏玩家异常退出后执行
func RctRemovePlayer(roomid string, palyerid string) {
	stream, err := rcclient.RemovePlayer(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	//这里后面加个chan，里面一有消息便开始发送
	for {
		data := <-Chan_RemovePlayer
		datas := strings.Split(data, ",")
		stream.Send(&ReqRemovePlayer{RoomId: datas[0], PlayId: datas[1]})
	}
}
