package main

import "paguma/pgnet"

func main() {
	// 1. 创建一个server句柄
	s := pgnet.NewServer("[paguma v0.1]")
	// 2. 启动服务
	s.Serve()
}

