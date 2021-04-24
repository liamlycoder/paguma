package main

import (
	"fmt"
	"paguma/pgiface"
	"paguma/pgnet"
)

// PingRouter ping test 自定义路由
type PingRouter struct {
	pgnet.BaseRouter
}


func (pr *PingRouter)Handle(request pgiface.IRequest)  {
	fmt.Println("Call Router Handle...")
	// 先读取客户端的数据，再写回ping
	fmt.Println("receive from client: msgID = ", request.GetMsgID(), ", data = ", string(request.GetData()))
	err := request.GetConnection().SendMsg(1, []byte("Hello, Liamcoder"))
	if err != nil {
		fmt.Println(err)
	}
}


func main() {
	// 1. 创建一个server句柄
	s := pgnet.NewServer("[paguma v0.5]")

	// 2. 添加一个自定义的router
	s.AddRouter(&PingRouter{})

	// 3. 启动服务
	s.Serve()
}

