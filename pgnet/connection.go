package pgnet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"paguma/pgiface"
	"paguma/utils"
	"sync"
)

/*
链接模块
*/

type Connection struct {
	// 当前Conn隶属于哪个server
	TcpServer pgiface.IServer

	// 当前链接的socket TCP套接字
	Conn *net.TCPConn

	// 当前链接ID
	ConnID uint32

	// 当前链接状态
	isClosed bool

	// 告知当前链接已经退出的/停止的 channel  (实际上是由Reader告知Writer退出的信号)
	ExitChan chan bool

	// 无缓冲管道，用于读写goroutine之间的消息通信。（思考：为什么这里不需要缓冲区？加了缓冲会怎样?）
	msgChan chan []byte

	// 消息的管理Msg和对应的API
	MsgHandler pgiface.IMsgHandler

	//链接属性
	property map[string]interface{}

	////保护当前property的锁
	propertyLock sync.Mutex
}

// NewConnection 初始化链接模块的方法
func NewConnection(server pgiface.IServer, conn *net.TCPConn, connID uint32, msgHandler pgiface.IMsgHandler) *Connection {
	c := &Connection{
		TcpServer:  server,
		Conn:       conn,
		ConnID:     connID,
		isClosed:   false,
		MsgHandler: msgHandler,
		ExitChan:   make(chan bool, 1),
		msgChan:    make(chan []byte),
		property:   make(map[string]interface{}),
	}

	// 将conn加入到ConnManager中
	c.TcpServer.GetConnMgr().Add(c)

	return c
}

// StartReader 链接的读业务方法
func (c *Connection) StartReader() {
	fmt.Println("【Reader Goroutine is running...】")
	defer fmt.Println("【Reader is exited】", "connID = ", c.ConnID, ", remote addr is ", c.RemoteAddr().String())
	defer c.Stop() // 当读业务出现任何异常，都会调用Stop函数，而Stop函数中可以将Reader退出的消息发送到ExitChan管道上去通知Writer也退出

	for {
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

		if utils.GlobalObject.WorkerPoolSize > 0 {
			// 已经开启了工作池机制，将消息发送给worker工作池
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			// 没有开启工作池，直接按原来的goroutine来处理
			// 从路由中找到注册绑定的Conn对应的router调用
			go c.MsgHandler.DoMsgHandler(&req)
		}

	}

}

// StartWriter 链接的写业务方法
func (c *Connection) StartWriter() {
	fmt.Println("【Writer Goroutine is running...】")
	defer fmt.Println("【Writer is exited】,", "connID = ", c.ConnID, "remote addr is ", c.RemoteAddr().String())

	// 不断的阻塞等待channel的消息，进行写给客户端
	for {
		select {
		case data := <-c.msgChan:
			// 有数据写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data error: ", err)
				return
			}
		case <-c.ExitChan:
			// 如果ExitChan可读了，说明此时Reader已退出，此时Writer也应该退出
			return
		}
	}

}

// Start 启动链接，让当前的链接准备开始工作
func (c *Connection) Start() {
	fmt.Println("Connection start...connID = ", c.ConnID)

	// 启动当前链接的读数据的业务
	go c.StartReader()
	// 启动当前链接的写数据的业务
	go c.StartWriter()
	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.TcpServer.CallOnConnStart(c)
}

// Stop 停止链接，结束当前链接的工作
func (c *Connection) Stop() {
	fmt.Println("Connection stopped...connID = ", c.ConnID)
	// 先判断是否已经关闭
	if c.isClosed == true {
		return
	}

	c.isClosed = true

	c.TcpServer.CallOnConnStop(c)

	// 关闭socket链接
	c.Conn.Close()

	// 告知Writer关闭
	c.ExitChan <- true

	// 将当前链接从ConnMgr中移除
	c.TcpServer.GetConnMgr().Remove(c)

	// 回收资源
	close(c.ExitChan)
	close(c.msgChan)
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
	// 将数据通过msgChan发送给Writer
	c.msgChan <- binaryMsg

	return nil
}

//SetProperty 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	if c.property == nil {
		c.property = make(map[string]interface{})
	}

	c.property[key] = value
}

//GetProperty 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	}

	return nil, errors.New("no property found")
}

//RemoveProperty 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
