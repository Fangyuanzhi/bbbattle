package main

import (
	"math"
	"sync"
)

const BORN_SIZE = 48

type Player struct {
	id uint64
	// colorId     uint32
	roomId         string
	name           string
	self           *PlayerTask
	pos            Vector
	weight         float32
	speed          float32
	scale          float32
	token          string
	isChangePos    bool
	isChangeWeight bool
	rotation       Vector //朝向
}
type SendPlayer struct {
	PlayerId uint64
	Name     string
	Pos      Vector
	Rotation Vector
	Weight   float32
	IsLocal  bool
}

func (p *Player) SetScale() {
	p.scale = float32(math.Sqrt(float64(p.weight))) / BORN_SIZE
}

var SyncPlayers = new(sync.Map)

func NewPlayer(t *PlayerTask) *Player {
	p := &Player{
		id:   t.playerId,
		name: t.account,
		//colorId:  0,
		pos:      Vector{(rand1.Float32()*2 - 1) * 15, (rand1.Float32()*2 - 1) * 8},
		rotation: Vector{rand1.Float32()*2 - 1, rand1.Float32()*2 - 1},
		weight:   256.0,
		speed:    1000,
		roomId:   t.room.roomId,
		token:    t.token,
		self:     t,
	}
	t.player = p
	return p
}
func (p *Player) Move(newpos Vector, rotation Vector) {
	if p.pos == newpos && p.rotation == rotation {
		return
	}
	p.pos = newpos
	p.rotation = rotation
	p.isChangePos = true
	//找到周围的目标

	// batchfood := &BatchMsgFood{}
	// for _, fid := range foodids {
	// 	food := room.foods.FoodList[fid]
	// 	batchfood.FoodList = append(batchfood.FoodList, food)
	// }
	// if len(batchfood.FoodList) == 0 {
	// 	return
	// }
	// sendData := MarshalData(int32(RoomCmd_SpawnFood), batchfood)
	// p.self.SenBuff(sendData)
}
func (p *Player) Stop() {
	p.pos = Vector{0, 0}
	p.rotation = Vector{0, 0}
	p.isChangePos = false
}
