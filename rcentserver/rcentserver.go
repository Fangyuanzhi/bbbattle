package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

type RCenterService struct{}

//var rct *RCenterService

func (r *RCenterService) GetRoomList(stream RCenter_GetRoomListServer) error {
	_, err := stream.Recv()
	if err != nil {
		log.Println(err)
	}
	stream.Send(&RetGetRoomList{RoomIds: redismgr.GetRoomList()})
	return nil
}
func (r *RCenterService) GetRoom(stream RCenter_GetRoomServer) error {
	// args, err := stream.Recv()
	// if err != nil {
	// 	log.Println(err)
	// }

	// stream.SendAndClose(&RetGetRoom{RoomId: args.RoomId, Addr: redismgr.GetRoom(args.GetRoomId())})
	// addr := "192.168.1.1"
	// reply := &RetGetRoom{RoomId: args.RoomId, Addr: addr}
	return nil
}

//
func (r *RCenterService) GetToken(stream RCenter_GetTokenServer) error {
	args, err := stream.Recv()
	if err != nil {
		log.Println(err)
	}
	name := args.GetName()
	stream.Send(&RetGetToken{Token: redismgr.GetToken(name)})
	return nil
}
func (r *RCenterService) DelRoom(stream RCenter_DelRoomServer) error {
	for {
		args, err := stream.Recv()
		if err != nil {
			log.Println(err)
			continue
		}
		redismgr.DelRoom(args.GetRoomId())
		// if ok := redismgr.DelRoom(args.RoomId); !ok {
		// 	continue
		// }
	}
}
func (r *RCenterService) UpdateRoomList(stream RCenter_UpdateRoomListServer) error {
	for {
		args, err := stream.Recv()
		if err != nil {
			log.Println(err)
			return err
		}
		//args.GetRooms()
		redismgr.UpdateRoomList(args.GetRoomIds(), args.GetNums())
		for {
			stream.Send(&RetUpdateRoomList{})
		}
	}
	//return nil
}
func (r *RCenterService) RemovePlayer(stream RCenter_RemovePlayerServer) error {
	for {
		args, err := stream.Recv()
		if err != nil {
			log.Println(err)
			return err
		}
		redismgr.RemovePlayer(args.GetRoomId(), args.GetPlayId())
	}

	//return nil
}
func (r *RCenterService) mustEmbedUnimplementedRCenterServer() {}
func GrpcStart() {
	grpcserver := grpc.NewServer()
	RegisterRCenterServer(grpcserver, new(RCenterService))
	lis, err := net.Listen("tcp", ":1235")
	if err != nil {
		log.Fatal(err)
	}
	grpcserver.Serve(lis)
}
