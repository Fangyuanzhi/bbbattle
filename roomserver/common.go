package main

import "math"

type SortNode struct {
	Key   uint64
	Value interface{}
}
type SortList struct {
	SList []SortNode
}
type Vector struct {
	X float32
	Y float32
}

type Command struct {
	Cmd  int32  `json:"cmd"`
	Data string `json:"data"`
}

func (v *Vector) ToMsg() *MsgVector {
	if v == nil {
		return nil
	}
	return &MsgVector{
		X: v.X,
		Y: v.Y,
	}
}
func (v Vector) Len() float64 {
	return math.Sqrt(float64(v.X)*float64(v.X) + float64(v.Y)*float64(v.Y))
}
func (v Vector) Sub(value *Vector) *Vector {
	return &Vector{v.X - value.X, v.Y - value.Y}
}
func (v Vector) Distance(o *Vector) float64 {
	subv := v.Sub(o)
	return subv.Len()
}
func (s *SortList) Insert(key uint64, value interface{}) {
	var (
		size = s.Size()
		low  = 0
		high = size - 1
	)
	if size <= 0 || key < s.SList[0].Key {
		s.SList = append([]SortNode{{key, value}}, s.SList...)
		return
	}
	mid := (low + high) >> 1
	for low < high {
		if key < s.SList[mid].Key {
			high = mid - 1
		} else {
			low = mid + 1
		}
		mid = (low + high) >> 1
		// else {
		// 	s.SList = append(append(s.SList[:mid], SortNode{key, value}), s.SList[mid:]...)
		// 	break
		// }
		// if low == high {
		// 	s.SList = append(append(s.SList[:low], SortNode{key, value}), s.SList[low:]...)
		// }
	}
	temp := append([]SortNode{}, s.SList[mid+1:]...)
	s.SList = append(append(s.SList[:mid+1], SortNode{key, value}), temp...)
}
func (s *SortList) Erase(key uint64) bool {
	var (
		size = s.Size()
		low  = 0
		high = size - 1
	)
	if size <= 0 || key < s.SList[0].Key {
		return false
	}
	mid := (low + high) >> 1
	for low < high {
		if key < s.SList[mid].Key {
			high = mid - 1
		} else {
			low = mid + 1
		}
		mid = (low + high) >> 1
	}
	if s.SList[mid].Key != key {
		return false
	}
	s.SList = append(s.SList[:mid], s.SList[mid+1:]...)
	return true
}
func (s *SortList) Size() int {
	return len(s.SList)
}
func (s *SortList) Find(key uint64) interface{} {
	var (
		size = s.Size()
		low  = 0
		high = size - 1
	)
	if size == 0 {
		return nil
	}
	mid := (low + high) >> 1
	for low < high {
		if key < s.SList[mid].Key {
			high = mid - 1
		} else if key > s.SList[mid].Key {
			low = mid + 1
		} else {
			break
		}
		mid = (low + high) >> 1
	}
	return s.SList[mid].Value
}
func (s *SortList) FindByIdx(idx int) interface{} {
	return s.SList[idx].Value
}
func (s *SortList) Clear() {
	s.SList = []SortNode{}
}
func (s *SortList) List() []SortNode {
	return s.SList
}

const (
	MaxCompressSize = 1024 * 1024
	CmdHeaderSize   = 4
)

type MsgFood1 struct {
}
