package main

import (
	"context"
	"log"

	redis "github.com/go-redis/redis/v8"
)

type RedisMgr struct{}

var redismgr *RedisMgr
var rdb *redis.Client
var ctx = context.Background()

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
func (r RedisMgr) GetName(userid string) string {
	name, err := rdb.HGet(ctx, "name", userid).Result()
	if err != nil {
		log.Println(err)
		return ""
	}
	return name
}

//简易登录验证
func (r *RedisMgr) CheckId(userid, password string) bool {
	pwd, err := rdb.HGet(ctx, "password", userid).Result()
	if err != nil {
		log.Printf("get password error:%v", err)
		return false
	}
	if pwd == password {
		return true
	}
	return false
}
func (r *RedisMgr) SetName(userid, name string) bool {
	_, err := rdb.HSet(ctx, "name", userid, name).Result()
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

//简易账号注册
func (r *RedisMgr) RegId(userid, password string) bool {
	ok, err := rdb.HExists(ctx, "password", userid).Result()
	//nums, err := rdb.Exists(ctx, userid+":password").Result()
	if err != nil {
		log.Printf("get keys nums error:%v", err)
		return false
	}
	if !ok {
		_, err := rdb.HSet(ctx, "password", userid, password).Result()
		if err != nil {
			log.Printf("set password error:%v", err)
			return false
		}
		return true
	}
	return false
}

//加入在线列表
func (r *RedisMgr) AddOnlineList(userid string) {
	_, err := rdb.SAdd(ctx, "allonlineplaylist").Result()
	if err != nil {
		log.Println("add online list error:", err)
	}
}

//移除在线列表
func (r *RedisMgr) DelOnlineList(userid string) {
	_, err := rdb.SRem(ctx, "allonlineplaylist").Result()
	if err != nil {
		log.Println("add online list error:", err)
	}
}

func (r *RedisMgr) Final() {
	rdb.Close()
}
