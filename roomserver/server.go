package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

type RCenterService struct{}

//var rct *RCenterService

func (r *RCenterService) GetRoomList(args RCenter_GetRoomListServer) error {
	return nil
}
func (r *RCenterService) GetRoom(args RCenter_GetRoomServer) error {
	// addr := "192.168.1.1"
	// reply := &RetGetRoom{RoomId: args.RoomId, Addr: addr}
	return nil
}

//reids里面读token暂时没写
func (r *RCenterService) GetToken(args RCenter_GetTokenServer) error {

	return nil
}
func (r *RCenterService) DelRoom(args RCenter_DelRoomServer) error {

	return nil
}
func (r *RCenterService) UpdateRoomList(args RCenter_UpdateRoomListServer) error {

	return nil
}
func (r *RCenterService) RemovePlayer(args RCenter_RemovePlayerServer) error {
	return nil
}
func (r *RCenterService) mustEmbedUnimplementedRCenterServer() {}
func GrpcStart() {
	grpcserver := grpc.NewServer()
	RegisterRCenterServer(grpcserver, new(RCenterService))
	lis, err := net.Listen("tcp:", ":2222")
	if err != nil {
		log.Fatal(err)
	}
	grpcserver.Serve(lis)
}
