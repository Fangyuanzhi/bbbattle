package main

import (
	"log"
)

//获取在线状态
func (r *RedisMgr) GetOnlineStatus(userid string) bool {
	str, err := rdb.Get(ctx, userid+"status").Result()
	if err != nil {
		log.Fatal("user not online,error:", err)
		return false
	}
	if str == "1" {
		return true
	}
	return false
}

//获取好友列表
func (r *RedisMgr) GetFriendList(userid string) (flist []string) {
	flist, err := rdb.SMembers(ctx, userid+"flist").Result()
	if err != nil {
		log.Println("get friends list error:", err)
	}
	return
}

//关注
func (r *RedisMgr) ASubB(userida, useridb string) bool {
	//此处应该是原子操作，暂时没写
	_, err := rdb.SAdd(ctx, userida+"sub", useridb).Result()
	if err != nil {
		log.Println("sadd failed error:", err)
		return false
	}
	_, err = rdb.SAdd(ctx, useridb+"fans", userida).Result()
	if err != nil {
		log.Println("sadd failed error:", err)
		return false
	}
	ok, err := rdb.SIsMember(ctx, useridb+"sub", userida).Result()
	if err != nil {
		log.Println("sim error:", err)
	}
	if ok {
		_, err := rdb.SAdd(ctx, userida+"flist", useridb).Result()
		if err != nil {
			log.Println("sadd friend error:", err)
		}
		_, err = rdb.SAdd(ctx, useridb+"flist", userida).Result()
		if err != nil {
			log.Println("sadd friend error:", err)
		}
	}
	return true
}

//取关
func (r *RedisMgr) AUnsubB(userida, useridb string) bool {
	_, err := rdb.SRem(ctx, userida+"sub", useridb).Result()
	if err != nil {
		log.Println("errer:", err)
	}
	rdb.SRem(ctx, userida+"flsit", useridb)
	rdb.SRem(ctx, useridb+"flsit", userida)
	return true
}

//获取关注列表
func (r *RedisMgr) GetSublist(userid string) (slist []string) {
	slist, err := rdb.SMembers(ctx, userid+"sub").Result()
	if err != nil {
		log.Println("get sub list error:", err)
	}
	return
}

//获取粉丝列表
func (r *RedisMgr) GetFanslist(userid string) (slist []string) {
	slist, err := rdb.SMembers(ctx, userid+"fans").Result()
	if err != nil {
		log.Println("get fans list error:", err)
	}
	return
}

//获取在线列表
func (r *RedisMgr) GetOnlineList() (online_list []string) {
	online_list, err := rdb.SMembers(ctx, "allonlineplaylist").Result()
	if err != nil {
		log.Println("get online player error:", err)
	}
	return
}
