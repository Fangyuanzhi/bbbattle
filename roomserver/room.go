package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	RoomMaxPlayer = 50
)

// type RoomPlayer struct {
// 	Roomid   string
// 	Playerid string
// }

var (
	Chan_DelRoom      = make(chan string, 4)
	Chan_StartRoom    = make(chan string, 4)
	Chan_RemovePlayer = make(chan string, 10)

	//RemainFood        = &BatchMsgFood{}
)

type Map struct {
	RoomId  string       `json:"roomId"`
	Length  float32      `json:"length"`
	Width   float32      `json:"width"`
	Players []SendPlayer `json:"players"`
	Foods   []*MsgFood   `json:"foods"`
}
type EatFood struct {
	FoodId   uint64 `json:"foodid"`
	PlayerId uint64 `json:"playerid"`
}
type EatPlayer struct {
	PlayerId   uint64 `json:"playerid"`
	BePlayerId uint64 `json:"beplayerid"`
}
type WinPlayer struct {
	PlayerId uint64 `json:"playerid"`
}
type Room struct {
	roomId          string
	maxNum          uint16
	minNum          uint16
	maxWeight       float32
	length          float32
	width           float32
	isStart         int32
	players         SortList
	foods           BatchMsgFood
	winnerid        uint64
	dnode           *DlinkNode
	startTime       time.Time
	chan_PlayerMove chan *PlayerMove
	chan_EatFood    chan *EatFood
	chan_EatPlayer  chan *EatPlayer
	chan_Control    chan int //用来控制房间循环结束的
	chan_Online     chan *PlayerTask
	chan_Offline    chan *PlayerTask
}

var Rooms = new(sync.Map)

//先直接手动分配几个房间用来测试
var rooms [4]Room

func CreateRooms() {
	for i := 0; i < 4; i++ {
		rooms[i] = Room{
			roomId:          "room" + strconv.Itoa(i+1),
			maxNum:          5,
			minNum:          2,
			length:          16,
			width:           16,
			isStart:         0,
			maxWeight:       0,
			chan_PlayerMove: make(chan *PlayerMove, RoomMaxPlayer),
			chan_Control:    make(chan int, 1),
			chan_EatFood:    make(chan *EatFood, 20),
			chan_EatPlayer:  make(chan *EatPlayer, 3),
		}
		rooms[i].dnode = &DlinkNode{}
		node := &DlinkNode{
			key:  "tail",
			xpre: rooms[i].dnode,
			ypre: rooms[i].dnode,
		}
		*rooms[i].dnode = DlinkNode{
			key:   "head",
			xnext: node,
			ynext: node,
		}
		Rooms.Store(rooms[i].roomId, &rooms[i])
	}
}
func NewRoom(tokenString string) *Room {
	_, claims, err := ParseToken(tokenString)
	if err != nil {
		log.Println(err)
	}
	room := &Room{
		roomId:          claims.RoomId,
		maxNum:          5,
		minNum:          2,
		length:          16,
		width:           16,
		isStart:         0,
		maxWeight:       0,
		chan_PlayerMove: make(chan *PlayerMove, RoomMaxPlayer),
		chan_Control:    make(chan int, 1),
		chan_EatFood:    make(chan *EatFood, 20),
		chan_EatPlayer:  make(chan *EatPlayer, 3),
	}
	room.dnode = &DlinkNode{}
	node := &DlinkNode{
		key:  "tail",
		xpre: room.dnode,
		ypre: room.dnode,
	}
	*room.dnode = DlinkNode{
		key:   "head",
		xnext: node,
		ynext: node,
	}
	return room
}
func (r *Room) MainLoop() {
	defer func() {
		r.isStart = 0
	}()
	for {
		select {
		case pmove := <-r.chan_PlayerMove:
			//fmt.Println("pmove:", pmove)
			r.Move(pmove.playerid, pmove.duration, pmove.rotation)
		case EatFood := <-r.chan_EatFood:
			//r.AoiEatFood(EatFood.PlayerId, EatFood.FoodId)
			r.EatFood(EatFood.PlayerId, EatFood.FoodId)
			//fmt.Printf("EatFood.FoodId: %v\n", EatFood.FoodId)
		case eatplayer := <-r.chan_EatPlayer:
			r.EatPlayer(eatplayer.PlayerId, eatplayer.BePlayerId)
		case playertask := <-r.chan_Offline:
			fmt.Println(playertask)
		case playertask := <-r.chan_Online:
			r.PlayerOnline(playertask)
		case <-r.chan_Control:
			return
		}
	}
}
func (r *Room) GetOnlinePlayerIds() (list []uint64) {
	for _, p := range r.players.SList {
		list = append(list, p.Key)
		//p.Value.(*Player).id
	}
	return
}
func (r *Room) PlayerOnline(p *PlayerTask) {
	player := &SendPlayer{
		PlayerId: p.playerId,
		Name:     p.player.name,
		Pos:      p.player.pos,
		Weight:   p.player.weight,
	}
	sendData := MarshalData(int32(RoomCmd_AddDelPlayer), player)
	r.SendCmdToAll(sendData)
}
func (r *Room) PlayerOffline(p *PlayerTask) {
	player := &SendPlayer{
		PlayerId: p.playerId,
	}
	sendData := MarshalData(int32(RoomCmd_AddDelPlayer), player)
	r.SendCmdToAll(sendData)
}
func (r *Room) AddPlayer(p *Player) {
	r.players.Insert(p.id, p)
}
func (r *Room) RemovePlayer(p *Player) {
	r.players.Erase(p.id)
	Chan_RemovePlayer <- r.roomId + "," + strconv.FormatUint(p.id, 10)
}
func (r *Room) GetPlayer(pid uint64) *Player {
	p, ok := r.players.Find(pid).(*Player)
	if !ok || p == nil {
		return nil
	}
	return p
}

// func (r *Room) Online(ptask PlayerTask) {
// 	p, isfind := r.players.Find(ptask.playerId).(*Player)
// 	if isfind && p != nil {
// 		p.id = ptask.playerId
// 		p.name = ptask.account
// 		ptask.player = p
// 	} else {
// 		p = NewPlayer(&ptask)
// 		r.AddPlayer(p)
// 	}
// 	ptask.room = r
// 	if p.token != ptask.token {
// 		p.token = ptask.token
//		p.self = ptask
// 	}
// }
// func (r *Room) Offline(ptask PlayerTask) {

// }

//游戏时间到应该释放所有资源
func (r *Room) EndTime() {

	time.Sleep(time.Minute * 1)
	time.Sleep(time.Second * 1)
	r.CloseRoom()
	r.Stop()
}
func (r *Room) Stop() {
	r.foods = BatchMsgFood{}
	r.isStart = 0
	r.maxWeight = 0
	fmt.Println(r.roomId + "游戏结束")
	//为避免心跳携程重复释放资源这里应当清除连接
	for _, v := range r.players.SList {
		conn := v.Value.(*Player).self.tcptask.Conn
		Conns.Delete(conn)
		conn.Close()
	}
	r.players.Clear()
	r.dnode = &DlinkNode{}
	node := &DlinkNode{
		key:  "tail",
		xpre: r.dnode,
		ypre: r.dnode,
	}
	*r.dnode = DlinkNode{
		key:   "head",
		xnext: node,
		ynext: node,
	}
	liveroom[r.roomId] = 2
	Chan_DelRoom <- r.roomId
	fmt.Println(r.roomId + "被放进通道里面")
	for {
		select {
		case <-r.chan_PlayerMove:
			fmt.Println("游戏结束了通道里的东西没读完，清一下")
		case <-r.chan_Control:
			fmt.Println("游戏结束了通道里的东西没读完，清一下")
		case <-r.chan_EatFood:
			fmt.Println("游戏结束了通道里的东西没读完，清一下")
		case <-r.chan_EatPlayer:
			fmt.Println("游戏结束了通道里的东西没读完，清一下")
		default:
			return
		}
	}
	// close(r.chan_Control)
	// close(r.chan_PlayerMove)
}
func (r *Room) Move(playid uint64, duration float32, rotation Vector) {
	// if pos.X < -15 || pos.Y < -8 || pos.X > 15 || pos.Y > 8 {
	// 	return
	// }
	player := r.GetPlayer(playid)
	if player == nil {
		fmt.Println("没找到")
		return
	}
	delta := &Vector{
		X: rotation.X * duration * Max(player.speed/player.weight, 0.5),
		Y: rotation.Y * duration * Max(player.speed/player.weight, 0.5),
	}
	newPos := &Vector{
		X: player.pos.X + delta.X,
		Y: player.pos.Y + delta.Y,
	}
	if newPos.X < -15 || newPos.Y < -8 || newPos.X > 15 || newPos.Y > 8 {
		return
	}
	player.Move(*newPos, rotation)
}
func (r *Room) Start() {
	r.isStart = 1
	r.startTime = time.Now()
}
func (r *Room) IsClosed() bool {
	return atomic.LoadInt32(&r.isStart) == 0
}
func (r *Room) CloseRoom() {
	go r.Control()
}

func (r *Room) End() {
	r.isStart = 0
	winner := &WinPlayer{
		PlayerId: r.winnerid,
	}
	sendData := MarshalData(int32(RoomCmd_End), &winner)
	r.SendCmdToAll(sendData)
}
func (r *Room) SendCmdToAll(data []byte) bool {
	// data, err := json.Marshal(cmd)
	// if err != nil {
	// 	log.Println("mashal failed", err)
	// }
	for _, pn := range r.players.SList {
		p, ok := pn.Value.(*Player)
		if !ok || p == nil {
			fmt.Println("没发出去")
			continue
		}
		p.self.SenBuff(data)
		//fmt.Printf("p.id: %v\n", p.id)
		//fmt.Println("发出去了")
	}
	return true
}

//这里后面设置成每隔指定时间执行一次
//移动
//这里加个aoi
func (r *Room) AoiMove() {
	//将自己向周围广播
	for _, pn := range r.players.SList {
		player, ok := pn.Value.(*Player)
		if !ok || player == nil {
			continue
		}
		if !player.isChangePos {
			continue
		}
		msgMove := &MsgMove{
			PlayerId: player.id,
			Pos:      player.pos.ToMsg(),
			Rotation: player.rotation.ToMsg(),
		}
		player.isChangePos = false
		//fmt.Println(player.pos)
		r.dnode.Remove(player.id)
		r.dnode.Insert(&DlinkNode{
			key: "player",
			val: player.id,
			x:   player.pos.X,
			y:   player.pos.Y,
		})
		sendData := MarshalData(int32(RoomCmd_Move), &msgMove)
		players := r.dnode.Find(player.id)
		//自己移动时周围向自己传位置
		// msgBatchMove := &MsgBatchMove{}
		// for _, pid := range players {
		// 	value := r.players.Find(pid)
		// 	if InterfaceIsNil(value) {
		// 		continue
		// 	}
		// 	p := value.(*Player)
		// 	if p.isChangePos {
		// 		continue
		// 	}
		// 	msgBatchMove.Moves = append(msgBatchMove.Moves, &MsgMove{
		// 		PlayerId: player.id,
		// 		Pos:      player.pos.ToMsg(),
		// 		Rotation: player.rotation.ToMsg(),
		// 	})
		// }
		// if len(msgBatchMove.Moves) != 0 {
		// 	sendData := MarshalData(int32(RoomCmd_BatchMove), msgBatchMove)
		// 	player.self.SenBuff(sendData)
		// }
		players = append(players, player.id)
		r.SendCmdToPlayers(players, sendData)
	}
}
func (r *Room) SyncMove() {
	msgMove := &MsgBatchMove{}
	for _, pn := range r.players.SList {
		player, ok := pn.Value.(*Player)
		if !ok || player == nil {
			continue
		}
		if !player.isChangePos {
			continue
		}
		player.isChangePos = false
		//players := r.dnode.Find(player.id)
		//r.SendCmdToPlayers(players, )
		msgMove.Moves = append(msgMove.Moves, &MsgMove{
			PlayerId: player.id,
			Pos:      player.pos.ToMsg(),
			Rotation: player.rotation.ToMsg(),
		})
		r.dnode.Remove(player.id)
		r.dnode.Insert(&DlinkNode{
			key: "player",
			val: player.id,
			x:   player.pos.X,
			y:   player.pos.Y,
		})
		//player.rotation.X = 0
		//player.rotation.Y = 0
	}
	if len(msgMove.Moves) == 0 {
		return
	}
	sendData := MarshalData(int32(RoomCmd_BatchMove), &msgMove)
	r.SendCmdToAll(sendData)
	//fmt.Println("q全员移动")
}
func (r *Room) MoveLoop() {
	for {
		if r.isStart == 0 {
			fmt.Println("游戏未开始")
			return
		}
		//r.SyncMove()
		r.AoiMove()
		time.Sleep(time.Millisecond * 10)
	}
}
func (r *Room) MoveNow(playerid uint64) {
	player := r.players.Find(playerid).(*Player)
	msgMove := &MsgMove{
		PlayerId: player.id,
		Pos:      player.pos.ToMsg(),
		Rotation: player.rotation.ToMsg(),
	}
	r.dnode.Remove(playerid)
	r.dnode.Insert(&DlinkNode{
		key: "playerid",
		val: playerid,
		x:   player.pos.X,
		y:   player.pos.Y,
	})
	//player.rotation.X = 0
	//player.rotation.Y = 0
	sendData := MarshalData(int32(RoomCmd_Move), &msgMove)
	r.SendCmdToAll(sendData)
}

//初始化设置房间信息以及食物位置
func (r *Room) Init() {
	for i := 0; i < 100; i++ {
		TFood := &MsgFood{}
		TFood.Id = uint64(i)
		TFood.Pos = &MsgVector{
			X: (rand1.Float32()*2 - 1) * 15,
			Y: (rand1.Float32()*2 - 1) * 8,
		}
		r.dnode.Insert(&DlinkNode{
			key: "food",
			val: TFood.Id,
			x:   TFood.Pos.X,
			y:   TFood.Pos.Y,
		})
		r.foods.FoodList = append(r.foods.FoodList, TFood)
	}
	log.Println("地图生成成功")
	gamemap := &Map{
		RoomId: r.roomId,
		Length: r.length,
		Width:  r.width,
	}
	gamemap.Players = make([]SendPlayer, r.players.Size())
	for i := 0; i < r.players.Size(); i++ {
		player := r.players.FindByIdx(i).(*Player)
		//fmt.Println(player)
		gamemap.Players[i] = SendPlayer{
			PlayerId: player.id,
			Name:     player.name,
			Pos:      player.pos,
			Rotation: player.rotation,
			Weight:   player.weight,
		}
		r.dnode.Insert(&DlinkNode{
			key: "player",
			val: player.id,
			x:   player.pos.X,
			y:   player.pos.Y,
		})
	}
	gamemap.Foods = make([]*MsgFood, len(r.foods.GetFoodList()))
	copy(gamemap.Foods, r.foods.GetFoodList())
	sendData := MarshalData(int32(RoomCmd_MapInit), &gamemap)

	r.SendCmdToAll(sendData)
	//地图初始化数据全部发送到玩家
	//继续发送初始玩家信息
	log.Println("地图初始化成功")
}

//吃完食物, 生成新食物数据
//改为仅全局更新体重信息
func (r *Room) UpdateWeight(playerid uint64) {
	player := r.players.Find(playerid).(*Player)
	msgweight := &MsgWeight{
		PlayerId: playerid,
		Weight:   player.weight,
	}
	sendData := MarshalData(int32(RoomCmd_UpdatePlayer), &msgweight)

	r.SendCmdToAll(sendData)

}
func (r *Room) EatFood(playerid uint64, foodid uint64) {
	foodpos := &Vector{
		X: r.foods.FoodList[foodid].Pos.X,
		Y: r.foods.FoodList[foodid].Pos.Y,
	}

	playerpos := r.players.Find(playerid).(*Player).pos
	if playerpos.Distance(foodpos) >= float64(r.players.Find(playerid).(*Player).scale)+0.1 {
		return
	}
	player := r.players.Find(playerid).(*Player)
	player.weight += 50
	player.isChangeWeight = true
	if player.weight > r.maxWeight {
		r.maxWeight, r.winnerid = player.weight, playerid
	}
	player.SetScale()
	r.foods.FoodList[foodid].Pos.X = (rand1.Float32()*2 - 1) * 15
	r.foods.FoodList[foodid].Pos.Y = (rand1.Float32()*2 - 1) * 8
	r.dnode.Remove(foodid)
	r.dnode.Insert(&DlinkNode{
		key: "food",
		val: foodid,
		x:   r.foods.FoodList[foodid].Pos.X,
		y:   r.foods.FoodList[foodid].Pos.Y,
	})
	eatfood := &EatFood{
		FoodId:   foodid,
		PlayerId: playerid,
	}
	//r.players.Find(playerid).(*Player).weight += 5.0
	sendData := MarshalData(int32(RoomCmd_EatFood), &eatfood)

	r.SendCmdToAll(sendData)

}

//aoi吃食物
func (r *Room) AoiEatFood(playerid uint64, foodid uint64) {

	foodpos := &Vector{
		X: r.foods.FoodList[foodid].Pos.X,
		Y: r.foods.FoodList[foodid].Pos.Y,
	}
	playerpos := r.players.Find(playerid).(*Player).pos
	if playerpos.Distance(foodpos) >= float64(r.players.Find(playerid).(*Player).scale)+0.1 {
		return
	}
	players := r.dnode.Find(foodid)
	fmt.Printf("players: %v\n", players)
	player := r.players.Find(playerid).(*Player)
	player.weight += 50
	player.isChangeWeight = true
	if player.weight > r.maxWeight {
		r.maxWeight, r.winnerid = player.weight, playerid
	}
	r.UpdateWeight(playerid)
	player.SetScale()
	r.foods.FoodList[foodid].Pos.X = (rand1.Float32()*2 - 1) * 15
	r.foods.FoodList[foodid].Pos.Y = (rand1.Float32()*2 - 1) * 8
	r.dnode.Remove(foodid)
	r.dnode.Insert(&DlinkNode{
		key: "food",
		val: foodid,
		x:   r.foods.FoodList[foodid].Pos.X,
		y:   r.foods.FoodList[foodid].Pos.Y,
	})
	food := &MsgFood{
		Id: foodid,
	}
	//r.players.Find(playerid).(*Player).weight += 5.0
	sendData := MarshalData(int32(RoomCmd_DelFood), &food)

	r.SendCmdToPlayers(players, sendData)
	//r.SendCmdToAll(sendData)
}

//这里加个aoi
func (r *Room) AoiSpawnFoodLoop() {
	for {
		if r.isStart == 0 {
			return
		}
		for _, pn := range r.players.SList {
			p, ok := pn.Value.(*Player)
			if !ok || p == nil {
				fmt.Println("没发出去")
				continue
			}
			batchfood := &BatchMsgFood{}
			for i := 0; i < 100; i++ {
				TFood := &MsgFood{}
				TFood.Id = uint64(i)
				TFood.Pos = &MsgVector{
					X: (rand1.Float32()*2 - 1) * 15,
					Y: (rand1.Float32()*2 - 1) * 8,
				}
				r.foods.FoodList = append(r.foods.FoodList, TFood)
			}
			foods := r.dnode.FindFood(p.id)
			for _, v := range foods {
				TFood := &MsgFood{}
				TFood.Id = v
				TFood.Pos = r.foods.FoodList[v].Pos
				batchfood.FoodList = append(batchfood.FoodList, TFood)
			}
			sendData := MarshalData(int32(RoomCmd_SpawnFood), &batchfood)
			p.self.SenBuff(sendData)
			//fmt.Printf("p.id: %v\n", p.id)
			//fmt.Println("发出去了")
		}

		time.Sleep(time.Second * 5)
	}
}
func (r *Room) SpawnFoodLoop() {
	for {
		if r.isStart == 0 {
			return
		}
		sendData := MarshalData(int32(RoomCmd_SpawnFood), &r.foods)

		r.SendCmdToAll(sendData)
		time.Sleep(time.Second * 5)
	}
}

func (r *Room) SendCmdToPlayers(players []uint64, data []byte) {
	for _, pid := range players {
		if r.isStart == 0 {
			return
		}
		playeri := r.players.Find(pid)
		if ok := InterfaceIsNil(playeri); ok {
			fmt.Println("inter nil")
			continue
		}
		player := playeri.(*Player)
		//fmt.Println(player.name)
		player.self.SenBuff(data)
	}
}
func (r *Room) EatPlayer(playerid uint64, beplayerid uint64) {

	player := r.players.Find(playerid).(*Player)

	beplayer := r.players.Find(beplayerid).(*Player)

	if player.weight <= beplayer.weight {
		return
	}
	// if player.weight == beplayer.weight || player.pos.Distance(&beplayer.pos) > float64(player.scale-beplayer.scale)+0.1 {
	// 	return
	// }
	player.weight += beplayer.weight
	if player.weight > r.maxWeight {
		r.maxWeight, r.winnerid = player.weight, playerid
	}
	fmt.Println("吃球了")
	beplayer.pos = Vector{
		X: (rand1.Float32()*2 - 1) * 15,
		Y: (rand1.Float32()*2 - 1) * 8,
	}
	r.MoveNow(beplayer.id)
	beplayer.weight = 256
	player.isChangeWeight = true
	beplayer.isChangeWeight = true
	sendData := MarshalData(int32(RoomCmd_EatPlayer), &EatPlayer{
		PlayerId:   player.id,
		BePlayerId: beplayer.id,
	})
	r.SendCmdToAll(sendData)
}

//告诉房间可以停止了
func (r *Room) Control() bool {
	if r.IsClosed() {
		return false
	}
	//Chan_DelRoom <- r.roomId
	r.chan_Control <- 1
	return true
}
