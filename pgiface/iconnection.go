package pgiface

import (
	"net"
)

// IConnection 定义链接模块的抽象层
type IConnection interface {
	Start()                   //启动连接，让当前连接开始工作
	Stop()                    //停止连接，结束当前连接状态M

	GetTCPConnection() *net.TCPConn //从当前连接获取原始的socket TCPConn
	GetConnID() uint32              //获取当前连接ID
	RemoteAddr() net.Addr           //获取远程客户端地址信息

	SendMsg(msgID uint32, data []byte) error     //直接将Message数据发送数据给远程的TCP客户端(无缓冲)

	SetProperty(key string, value interface{})   //设置链接属性
	GetProperty(key string) (interface{}, error) //获取链接属性
	RemoveProperty(key string)                   //移除链接属性
}

// HandleFunc 定义一个处理链接业务的方法
type HandleFunc func(*net.TCPConn, []byte, int) error
