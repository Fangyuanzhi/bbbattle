package main

import (
	"log"
	"sync"
)

var Aoi = new(sync.Map)

type DlinkNode struct {
	key                      string
	val                      uint64
	x, y                     float32
	xpre, xnext, ypre, ynext *DlinkNode
}

func (d *DlinkNode) Insert(dnode *DlinkNode) {
	cur := d
	Aoi.Store(dnode.val, dnode)
	//Aoi[dnode.val] = dnode
	for {
		if cur.xnext.key == "tail" || cur.xnext.x > dnode.x {
			dnode.xpre = cur
			dnode.xnext = cur.xnext
			cur.xnext.xpre = dnode
			cur.xnext = dnode
			break
		}
		cur = cur.xnext
	}
	cur = d
	for {
		if cur.ynext.key == "tail" || cur.ynext.y > dnode.y {
			dnode.ypre = cur
			dnode.ynext = cur.ynext
			cur.ynext.ypre = dnode
			cur.ynext = dnode
			break
		}
		cur = cur.ynext
	}
}
func (d *DlinkNode) Remove(key uint64) {
	value, ok := Aoi.Load(key)
	if !ok || InterfaceIsNil(value) {
		return
	}
	dnode := value.(*DlinkNode)
	// //dnode, ok := Aoi[key]
	// if !ok {
	// 	return
	// }
	dnode.xpre.xnext = dnode.xnext
	dnode.xnext.xpre = dnode.xpre
	dnode.ypre.ynext = dnode.ynext
	dnode.ynext.ypre = dnode.ypre
	Aoi.Delete(key)
	//delete(Aoi, key)
}
func (d *DlinkNode) Find(key uint64) (player []uint64) {
	value, ok := Aoi.Load(key)
	if !ok || InterfaceIsNil(value) {
		return
	}
	dnode := value.(*DlinkNode)
	if !ok {
		log.Println("aoi no find")
		return
	}
	//fmt.Printf("dnode: %v\n", dnode)
	cur := dnode.xpre
	for cur.key != "head" && dnode.x-cur.x < 8.2 {
		//fmt.Println(cur.y - dnode.y)
		if cur.key == "player" && cur.y-dnode.y < 6 && cur.y-dnode.y > -6 {
			player = append(player, cur.val)
			//fmt.Printf("cur: %v\n", cur)
			//fmt.Printf("dnode: %v\n", dnode)
		}
		cur = cur.xpre
	}
	cur = dnode.xnext
	for cur.key != "tail" && cur.x-dnode.x < 8.2 {
		if cur.key == "player" && cur.y-dnode.y < 6 && cur.y-dnode.y > -6 {
			player = append(player, cur.val)
			//fmt.Printf("cur: %v\n", cur)
			//fmt.Printf("dnode: %v\n", dnode)
		}
		cur = cur.xnext
	}
	return
}
func (d *DlinkNode) FindFood(key uint64) (food []uint64) {

	value, ok := Aoi.Load(key)
	if !ok || InterfaceIsNil(value) {
		return
	}
	dnode := value.(*DlinkNode)
	if !ok {
		log.Println("aoi no find")
		return
	}
	//fmt.Printf("dnode: %v\n", dnode)
	cur := dnode.xpre
	for cur.key != "head" && dnode.x-cur.x < 9 {
		//fmt.Println(cur.y - dnode.y)
		if cur.key == "food" && cur.y-dnode.y < 6 && cur.y-dnode.y > -6 {
			food = append(food, cur.val)
			//fmt.Printf("cur: %v\n", cur)
			//fmt.Printf("dnode: %v\n", dnode)
		}
		cur = cur.xpre
	}
	cur = dnode.xnext
	for cur.key != "tail" && cur.x-dnode.x < 9 {
		if cur.key == "food" && cur.y-dnode.y < 6 && cur.y-dnode.y > -6 {
			food = append(food, cur.val)
			//fmt.Printf("cur: %v\n", cur)
			//fmt.Printf("dnode: %v\n", dnode)
		}
		cur = cur.xnext
	}
	return
}
