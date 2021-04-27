package utils

import (
	"encoding/json"
	"io/ioutil"
	"paguma/pgiface"
)

/*
存储一切有关paguma框架的全局参数，供其他模块使用
一些参数是可以由用户的application.json配置的
 */

type GlobalObj struct {
	/*
		Server
	*/
	TCPServer pgiface.IServer //当前Paguma的全局Server对象
	Host      string         //当前服务器主机IP
	TCPPort   int            //当前服务器主机监听端口号
	Name      string         //当前服务器名称

	/*
		Paguma
	*/
	Version          string //当前Paguma版本号
	MaxPacketSize    uint32 //都需数据包的最大值
	MaxConn          int    //当前服务器主机允许的最大链接个数
	WorkerPoolSize   uint32   // 当前业务工作Worker池的goroutine数量
	MaxWorkerTaskLen uint32  // 表示框架允许开辟的最大Worker
}

//GlobalObject 全局对外的GlobalObj
var GlobalObject *GlobalObj

// 用来初始化当前GlobalObj对象
func init()  {
	//初始化GlobalObject变量，设置一些默认值
	GlobalObject = &GlobalObj{
		Host:          "0.0.0.0",
		TCPPort:       8999,
		Name:          "PagumaServer",
		Version:       "v0.9",
		MaxPacketSize: 4096,
		MaxConn:       1000,
		WorkerPoolSize: 10,
		MaxWorkerTaskLen: 1024,  // 这个需要写死
	}

	// 尝试从配置文件中加载配置参数
	GlobalObject.Reload()
}

// Reload 从application.json去加载，用于自定义参数
func (g *GlobalObj)Reload()  {
	data, err := ioutil.ReadFile("config/application.json")
	if err != nil {
		panic(err)
	}
	// 将json文件解析到struct中
	if err = json.Unmarshal(data, &GlobalObject); err != nil {
		panic(err)
	}
}
