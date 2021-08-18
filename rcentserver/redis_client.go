package main

import (
	"log"

	redis "github.com/go-redis/redis/v8"
)

type RedisMgr struct{}

var redismgr *RedisMgr
var rdb *redis.Client

func (r *RedisMgr) Init() bool {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "192.168.96.242:6379",
		Password: "",
		DB:       0,
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("redis connect failed, error:%v", err)
		return false
	}
	return true
}
func (r *RedisMgr) GetToken(name string) string {
	token, err := rdb.HGet(ctx, "token", name).Result()
	if err != nil {
		log.Println("err")
		return ""
	}
	return token
}
func (r *RedisMgr) SetToken(name string, token string) bool {
	//fmt.Println(name)
	_, err := rdb.HSet(ctx, "token", name, token).Result()
	if err != nil {
		log.Println("err")
		return false
	}
	return true
}
func (r *RedisMgr) GetRoom(roomid string) string {
	addr, err := rdb.Get(ctx, roomid+":roominfo").Result()
	if err != nil {
		log.Println(err)
		return ""
	}
	return addr
}
func (r *RedisMgr) DelRoom(roomid string) bool {
	_, err := rdb.ZAdd(ctx, "room", &redis.Z{Score: 5, Member: roomid}).Result()
	if err != nil {
		log.Println(err)
		return false
	}
	return false
}
func (r *RedisMgr) GetRoomList() (list []string) {
	list, err := rdb.ZRange(ctx, "room", 0, -1).Result()
	if err != nil {
		log.Println(err)
		return
	}
	return
}
func (r *RedisMgr) UpdateRoomList(roomids []string, nums []int32) bool {
	//data := make([]redis.Z, 0)
	data := []*redis.Z{}
	for i := 0; i < len(roomids); i++ {
		data = append(data, &redis.Z{Score: float64(nums[i]), Member: roomids[i]})
	}
	_, err := rdb.ZAdd(ctx, "room", data...).Result()
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
func (r *RedisMgr) RemovePlayer(roomid string, playerid string) bool {
	_, err := rdb.ZIncrBy(ctx, "room", float64(1), roomid).Result()
	if err != nil {
		log.Println(err)
		return false
	}
	//这里移除掉token后玩家便无法重连
	_, err = rdb.HDel(ctx, "token", playerid).Result()
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
