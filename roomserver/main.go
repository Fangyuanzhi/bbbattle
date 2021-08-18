package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	reflect "reflect"
	"strconv"
	sync "sync"
	"time"
	"unsafe"
)

func Max(x, y float32) float32 {
	if x > y {
		return x
	}
	return y
}

type HeartClient struct {
	playerid       uint64
	lastUpdatetime time.Time
}
type RoomServer struct {
	//id       string
	listener net.Listener
}

var rand1 = rand.New(rand.NewSource(time.Now().UnixNano()))

//管理所有连接统一处理心跳
var Conns = new(sync.Map)

//暂时没用
var PlayerTasks = make(map[string]*PlayerTask)
var ctx context.Context

// var roomserver *RoomServer
// var roomserver1 *RoomServer
// var roomserver2 *RoomServer
// var roomserver3 *RoomServer

var roomserver [4]RoomServer

// var roomCmd *RoomCmd

// func bytes2str(b []byte) string {
// 	return *(*string)(unsafe.Pointer(&b))
// }
func InterfaceIsNil(i interface{}) bool {
	ret := i == nil
	if !ret {
		defer func() {
			recover()
		}()
		ret = reflect.ValueOf(i).IsNil()
	}
	return ret

}
func str2uint64(str string) (num uint64) {
	num = 0
	level := uint64(1)
	bytes := str2bytes(str)
	for i := len(bytes) - 1; i >= 0; i-- {
		num += uint64(bytes[i]-'0') * level
		level *= 10
	}
	return
}
func str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s)) // 获取s的起始地址开始后的两个 uintptr 指针
	h := [3]uintptr{x[0], x[1], x[1]}      // 构造三个指针数组
	return *(*[]byte)(unsafe.Pointer(&h))
}

//处理心跳（老）
// func process(conn net.Conn, conns *sync.Map) {
// 	defer conn.Close()
// 	for {
// 		reader := bufio.NewReader(conn)
// 		var buf [128]byte
// 		n, err := reader.Read(buf[:])
// 		if err != nil {
// 			log.Println(err)
// 			break
// 		}
// 		conns.Store(conn, client{
// 			lastUpdatetime: time.Now(),
// 		})
// 		recvStr := string(buf[:n])
// 		fmt.Println(recvStr)
// 	}
// }
func (r *RoomServer) MainLoop() {
	for {
		conn, err := r.listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go NewPlayerTaskTcp(conn).RecvLoop()
		// go func() {
		// 	Conns.Store(conn, &HeartClient{
		// 		lastUpdatetime: time.Now(),
		// 	})
		// }()
	}

}

// func (r *RoomServer) Init() {

// }
// func matchProcess(conn net.Conn) {
// 	defer conn.Close()
// 	player := new(PlayerTask)
// 	reader := bufio.NewReader(conn)
// 	for {
// 		msg, err := Decode(reader)
// 		if err == io.EOF {
// 			continue
// 		}
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}
// 		player.ParseMsg(msg)
// 	}
// }
func scan(conns *sync.Map) {
	for {
		time.Sleep(time.Duration(10) * time.Second)
		log.Println("start scan conn....")
		conns.Range(func(key, value interface{}) bool {
			client := value.(*HeartClient)
			if time.Since(client.lastUpdatetime).Seconds() > float64(10) {
				key.(net.Conn).Close()
				//下面通知相关资源释放
				player, _ := SyncPlayers.Load(client.playerid)
				player.(*Player).self.room.RemovePlayer(player.(*Player))
				player.(*Player).self.isLogin = false
				liveroom[player.(*Player).self.room.roomId]++
				conns.Delete(key)
				log.Println(key, "kick off")
			}
			return true
		})

	}
}
func main() {
	RCenterClientInit()
	CreateRooms()
	for i := 0; i < 4; i++ {
		listener, err := net.Listen("tcp", ":"+strconv.Itoa(i+40001))
		if err != nil {
			log.Fatal(err)
		}
		roomserver[i].listener = listener
		go roomserver[i].MainLoop()
	}
	//listener, err := net.Listen("tcp", ":1234")

	// roomserver.listener = listener
	// roomserver.MainLoop()
	go scan(Conns)
	go RctDelRoom()
	for {
		startroom := <-Chan_StartRoom
		room, _ := Rooms.Load(startroom)
		r := room.(*Room)
		r.isStart = 1
		r.startTime = time.Now()
		r.Init()
		go r.MainLoop()
		go r.MoveLoop()
		go r.SpawnFoodLoop()
		//go r.AoiSpawnFoodLoop()
		go r.EndTime()
	}
	// for {
	// 	conn, err := listener.Accept()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	//go process(conn, conns)
	// 	go matchProcess(conn)
	// }
}
