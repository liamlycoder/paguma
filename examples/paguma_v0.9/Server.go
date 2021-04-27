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
	err := request.GetConnection().SendMsg(200, []byte("ping...ping...ping..."))
	if err != nil {
		fmt.Println(err)
	}
}

// HelloRouter test 自定义路由
type HelloRouter struct {
	pgnet.BaseRouter
}


func (h *HelloRouter)Handle(request pgiface.IRequest)  {
	fmt.Println("Call Router Handle...")
	// 先读取客户端的数据，再写回ping
	fmt.Println("receive from client: msgID = ", request.GetMsgID(), ", data = ", string(request.GetData()))
	err := request.GetConnection().SendMsg(201, []byte("Hello, Liamcoder"))
	if err != nil {
		fmt.Println(err)
	}
}

// 创建链接之后执行的钩子函数
func begin(conn pgiface.IConnection)  {
	fmt.Println("===》DO CONNECTION BEGIN...")
	_ = conn.SendMsg(301, []byte("china telecom is garbage"))
}


// 销毁链接之前执行的钩子函数
func end(conn pgiface.IConnection)  {
	fmt.Println("===》DO CONNECTION END...")
	_ = conn.SendMsg(302, []byte("I must leave china telecom"))
}


func main() {
	// 1. 创建一个server句柄
	s := pgnet.NewServer()

	// 2. 注册钩子函数
	s.SetOnConnStart(begin)
	s.SetOnConnStop(end)

	// 3. 添加一个自定义的router
	s.AddRouter(0, &HelloRouter{})
	s.AddRouter(1, &PingRouter{})

	// 4. 启动服务
	s.Serve()
}

