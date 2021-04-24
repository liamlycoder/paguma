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

// 重写BaseRouter的三个方法（可以不全部重写，根据自己业务需求进行选择)

func (pr *PingRouter)PreHandle(request pgiface.IRequest)  {
	fmt.Println("Call Router PreHandle...")
	if _, err := request.GetConnection().GetTCPConnection().Write([]byte("before ping...\n")); err != nil {
		fmt.Println("call back before ping err: ", err)
	}
}

func (pr *PingRouter)Handle(request pgiface.IRequest)  {
	fmt.Println("Call Router Handle...")
	if _, err := request.GetConnection().GetTCPConnection().Write([]byte("ping...ping...ping...\n")); err != nil {
		fmt.Println("call back ping err: ", err)
	}
}

func (pr *PingRouter)PostHandle(request pgiface.IRequest)  {
	fmt.Println("Call Router PostHandle...")
	if _, err := request.GetConnection().GetTCPConnection().Write([]byte("after ping...\n")); err != nil {
		fmt.Println("call back after ping err: ", err)
	}
}

func main() {
	// 1. 创建一个server句柄
	s := pgnet.NewServer("[paguma v0.3]")

	// 2. 添加一个自定义的router
	s.AddRouter(&PingRouter{})

	// 3. 启动服务
	s.Serve()
}

