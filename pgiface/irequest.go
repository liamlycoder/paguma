package pgiface

/*
IRequest接口：
实际上是把客户端的【链接请求】和【请求数据】绑定到了一起
 */

type IRequest interface {
	// GetConnection 得到当前链接
	GetConnection() IConnection

	// GetData 得到请求的消息数据
	GetData() []byte
}
