package pgnet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"paguma/pgiface"
)

/*
链接模块
*/

type Connection struct {
	// 当前链接的socket TCP套接字
	Conn *net.TCPConn

	// 当前链接ID
	ConnID uint32

	// 当前链接状态
	isClosed bool

	// 告知当前链接已经退出的/停止的 channel
	ExitChan chan bool

	// 该链接处理的方法Router
	Router pgiface.IRouter
}

// NewConnection 初始化链接模块的方法
func NewConnection(conn *net.TCPConn, connID uint32, router pgiface.IRouter) *Connection {
	c := &Connection{
		Conn:     conn,
		ConnID:   connID,
		isClosed: false,
		Router:   router,
		ExitChan: make(chan bool, 1),
	}
	return c
}

// StartReader 链接的读业务方法
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running...")
	defer fmt.Println("connID = ", c.ConnID, " Reader is exit, remote addr is ", c.RemoteAddr().String())
	defer c.Stop()

	for {
		// 读取客户端的数据到buf中，其大小由配置文件指定
		//buf := make([]byte, utils.GlobalObject.MaxPacketSize)
		//_, err := c.Conn.Read(buf)
		//if err != nil {
		//	fmt.Println("recv buf err: ", err)
		//	continue
		//}
		// -------以上代码在拆包功能实现后被废除----------

		// 创建一个拆包、解包的对象
		dp := NewDataPack()
		// 读取客户端的Msg Head (二进制流，8字节)
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error: ", err)
			break
		}
		// 拆包，得到msgID 和 msgDataLen 放在msg消息中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error: ", err)
			break
		}
		// 根据dataLen 再次读取data，放在msg.data中
		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error: ", err)
				break
			}
		}
		msg.SetData(data)

		// 得到当前链接conn的request数据
		req := Request{
			conn: c,
			msg:  msg,
		}

		// 从路由中找到注册绑定的Conn对应的router调用
		go func(request pgiface.IRequest) {
			c.Router.PreHandle(request)
			c.Router.Handle(request)
			c.Router.PostHandle(request)
		}(&req)

	}

}


// Start 启动链接，让当前的链接准备开始工作
func (c *Connection) Start() {
	fmt.Println("Connection start...connID = ", c.ConnID)

	// 启动当前链接的读数据的业务
	go c.StartReader()

	// TODO 启动从当前链接写数据的业务

}

// Stop 停止链接，结束当前链接的工作
func (c *Connection) Stop() {
	fmt.Println("Connection stopped...connID = ", c.ConnID)
	// 先判断是否已经关闭
	if c.isClosed == true {
		return
	}

	c.isClosed = true

	// 关闭socket链接
	c.Conn.Close()

	// 回收资源
	close(c.ExitChan)
}

// GetTCPConnection 获取当前链接所绑定的socket conn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// GetConnID 获取当前链接模块的链接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// RemoteAddr 获取远程客户端的TCP状态（IP和端口）
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// SendMsg Send 发送数据，将数据发送给远程的客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed {
		return errors.New("connection closed when send msg")
	}
	// 将data进行封包，格式：MsgDataLen|MsgID|Data
	dp := NewDataPack()
	binaryMsg, err := dp.Pack(NewMsgPack(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("pack error msg")
	}
	// 将数据发送给客户端
	if _, err := c.Conn.Write(binaryMsg); err != nil {
		fmt.Println("write msg id: ", msgId, " error: ", err)
		return errors.New("conn write error")
	}
	return nil
}
