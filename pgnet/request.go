package pgnet

import "paguma/pgiface"

type Request struct {
	// 已经和客户端建立好的链接
	conn pgiface.IConnection

	// 客户端请求的数据
	msg pgiface.IMessage
}

// GetConnection 得到当前链接
func (r *Request)GetConnection() pgiface.IConnection {
	return r.conn
}

// GetData 得到请求的消息数据
func (r *Request)GetData() []byte {
	return r.msg.GetData()
}

func (r *Request)GetMsgID() uint32 {
	return r.msg.GetMsgId()
}
