package main

func main() {
	redismgr.Init()
	go RpcHandle()
	go GrpcStart()
	select {}
}
