package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type TcpTask struct {
	Conn     net.Conn
	StopChan chan struct{}
}
type PlayerTask struct {
	tcptask  *TcpTask
	playerId uint64
	account  string
	room     *Room
	player   *Player
	token    string
	//activeTime time.Time
	isLogin bool
}
type PlayerMove struct {
	playerid uint64
	pos      Vector
	rotation Vector
	duration float32
}

// var Cmd Command

// func NewTcpTask(conn net.Conn)*TcpTask{

// }
func NewPlayerTaskTcp(conn net.Conn) *PlayerTask {
	p := &PlayerTask{
		tcptask: &TcpTask{
			Conn:     conn,
			StopChan: make(chan struct{}),
		},
		//activeTime: time.Now(),
	}
	return p
}

// func (p *PlayerTask) Start() {
// 	go RecvLoop(p.tcptask.Conn, p)
// }
func (p *PlayerTask) RecvLoop() {
	for {
		reader := bufio.NewReader(p.tcptask.Conn)
		msg, err := Decode(reader)
		//读完可以退出了,相关协程关闭，资源释放交给心跳解决
		if err == io.EOF {
			//log.Println("总之是断开连接了，退出吧")
			return
		}
		if err != nil {
			log.Println(err)
			return
		}
		go p.ParseMsg(msg)
	}
}

// func RecvLoop(conn net.Conn, playertask *PlayerTask) {
// 	for {
// 		reader := bufio.NewReader(conn)
// 		msg, err := Decode(reader)
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}
// 		go playertask.ParseMsg(msg)
// 	}
// }

// func SendLoop(playertask *PlayerTask) {
// 	for {

// 	}
// }
func (p *PlayerTask) Stop() {
	Conns.Delete(p.tcptask.Conn)
	p.tcptask.Conn.Close()
	p.isLogin = false
}
func (p *PlayerTask) SenBuff(data []byte) bool {
	sendData, err := Encode(string(data))
	if err != nil {
		log.Println(err)
	}
	_, err = p.tcptask.Conn.Write(sendData)
	if err != nil {
		log.Println("tcp send error:", err)
		return false
	}
	return true
}
func (p *PlayerTask) Verified() {
	p.isLogin = true
}
func (p *PlayerTask) OnACtive() {
	//p.activeTime = time.Now()
}
func (p *PlayerTask) ParseMsg(data []byte) bool {
	// if len(data) < CmdHeaderSize {
	// 	return true
	// }
	Cmd := Command{}
	err := json.Unmarshal(data, &Cmd)
	if err != nil {
		log.Println("unmashal failed , error:", err)
		return false
	}
	cmd := Cmd.Cmd
	//fmt.Println(Cmd)
	if !p.isLogin {
		if cmd != int32(RoomCmd_Login) {
			log.Println("不是登录指令", err)
			return false
		}
		revCmd := &ReqRoomLogin{}
		err = json.Unmarshal(str2bytes(Cmd.Data), revCmd)
		if err != nil {
			log.Println("cmd unmashal failed error :", err)
			return false
		}
		//用grpc调用rcenterserver检查token
		ttoken, err := RctGettoken(revCmd.Name)
		if err != nil {
			log.Println("grpc 调用失败, error:", err)
		}
		if revCmd.GetToken() == ttoken.GetToken() {
			log.Println("登录成功")
			//下面一堆操作
			//存入token能直接获得的信息
			p.account = revCmd.Name
			p.token = revCmd.Token
			p.playerId = str2uint64(Getting(ttoken.GetToken()).UserId)
			fmt.Println(p.playerId)
			//存入房间相关信息
			room, _ := Rooms.Load(Getting(p.token).RoomId)
			//fmt.Println(room)
			p.room, _ = room.(*Room)
			value, ok := SyncPlayers.Load(p.playerId)
			if !ok {
				//fmt.Println(room.(*Room).roomId)
				//存入用户相关信息
				player := NewPlayer(p)
				//fmt.Println(player.name)
				player.SetScale()
				p.player = player
				p.room.players.Insert(p.playerId, p.player)
				SyncPlayers.Store(p.playerId, player)
				Conns.Store(p.tcptask.Conn, &HeartClient{
					lastUpdatetime: time.Now(),
					playerid:       p.playerId,
				})
				liveroom[p.room.roomId]--
				//新用户过来需要判断一下这个room是否能开始
				//fmt.Println(p.room.players.Size())
				p.Verified()
				//SyncPlayers.Store(p.playerId, p.player)
				if p.room.minNum <= uint16(p.room.players.Size()) {
					//游戏开始
					fmt.Println("游戏开始")
					Chan_StartRoom <- p.room.roomId
				}
			} else {
				if InterfaceIsNil(value) {
					return false
				}
				player := value.(*Player)
				p.player = player
				p.room.players.Insert(p.playerId, p.player)
				p.Verified()
			}
			// p.player = &Player{
			// 	id:     p.playerId,
			// 	roomId: Getting(p.token).RoomId,
			// 	name:   p.account,
			// 	self:   p,
			// 	token:  p.token,
			// }
			// p.room = &Room{
			// 	roomId: Getting(p.token).RoomId,
			// }
		} else {
			log.Println("登录验证失败:token不一样")
			return false
		}

		return true
	}
	//心跳
	if cmd == int32(RoomCmd_HeartBeat) {
		//更新心跳信息
		Conns.Store(p.tcptask.Conn, &HeartClient{
			lastUpdatetime: time.Now(),
			playerid:       p.playerId,
		})
		fmt.Println(p.account, "跳了一下")
		//p.activeTime = time.Now()
		return true
	}
	if p.room == nil {
		return false
	}
	switch cmd {
	case int32(RoomCmd_Move):
		revCmd := &MsgMove{}
		//fmt.Println(Cmd.Data)
		err = json.Unmarshal(str2bytes(Cmd.Data), revCmd)
		if err != nil {
			log.Println("move massage unmashal failed error:", err)
		}
		//fmt.Println("移动指令", revCmd.Rotation.X, revCmd.Rotation.Y)
		//判断移动是否正常已经其他不知道的操作,
		//fmt.Printf("%v %v", revCmd.Rotation.X, revCmd.Rotation.Y)
		//fmt.Printf("%v %v", revCmd.Pos.X, revCmd.Pos.Y)
		//更新位置,扔到chan里让某goroutine处理
		p.room.chan_PlayerMove <- &PlayerMove{
			playerid: p.playerId,
			pos: Vector{
				X: revCmd.Pos.GetX(),
				Y: revCmd.Pos.GetY(),
			},
			rotation: Vector{
				X: revCmd.Rotation.GetX(),
				Y: revCmd.Rotation.GetY(),
			},
			duration: revCmd.Duration,
		}
	case int32(RoomCmd_EatFood):
		revCmd := &EatFood{}
		err = json.Unmarshal(str2bytes(Cmd.Data), revCmd)
		if err != nil {
			log.Println("eat massage unmashal failed error:", err)
		}
		p.room.chan_EatFood <- &EatFood{
			PlayerId: p.playerId,
			FoodId:   revCmd.FoodId,
		}
	case int32(RoomCmd_EatPlayer):

		revCmd := &EatPlayer{}
		err = json.Unmarshal(str2bytes(Cmd.Data), revCmd)
		if err != nil {
			log.Println("eat massage unmashal failed error:", err)
		}
		if revCmd.PlayerId == p.playerId {
			p.room.chan_EatPlayer <- revCmd
		}
	case int32(RoomCmd_End):
		if p.room.isStart != 0 {
			if time.Since(p.room.startTime).Seconds() > float64(59) {
				p.room.End()
			}
		}
	default:
		//未定义的命令，直接忽略
		//直接扔某个chan里面让某个goroutine去解决
	}
	return true
}
