package pgnet

import (
	"fmt"
	"net"
	"paguma/pgiface"
	"paguma/utils"
)

type Server struct {
	// 服务器的名称
	Name string
	// 服务器绑定的ip版本
	IPVersion string
	// 服务器监听的ip
	IP string
	// 服务器监听的端口
	Port int
	// 当前Server的消息管理模块，用来绑定MsgID和对应处理业务的API
	MsgHandler pgiface.IMsgHandler
	// 该server的链接管理器
	ConnMgr pgiface.IConnManager
	//该Server的连接创建时Hook函数
	OnConnStart func(conn pgiface.IConnection)
	//该Server的连接断开时的Hook函数
	OnConnStop func(conn pgiface.IConnection)
}

func (s *Server) Start() {
	fmt.Printf("[Paguma] Server Name: %s, listener IP at : %s, Port : %d is starting...\n", utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TCPPort)
	fmt.Printf("[Paguma] Version %s, MaxConn: %d, MaxPacketSize: %d\n", utils.GlobalObject.Version, utils.GlobalObject.MaxConn, utils.GlobalObject.MaxPacketSize)
	go func() {
		// 0. 开启消息队列及工作池
		s.MsgHandler.StartWorkerPool()
		// 1. 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr error: ", err)
			return
		}
		// 2. 监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, " err: ", err)
			return
		}
		//已经监听成功
		fmt.Println("start Paguma server  ", s.Name, " succeed, now listening...")
		var cid uint32
		cid = 0
		// 3. 阻塞的等待客户端连接，处理客户端连接业务（读写）
		for {
			//3.1 阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err ", err)
				continue
			}
			fmt.Println("Get conn remote addr = ", conn.RemoteAddr().String())

			// 设置最大链接个数的判断，如果超过最大连接数，那么关闭此新的链接
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				// TODO 给客户端响应一个超出最大链接的错误包
				fmt.Println("Too many connections, MaxConnection = ", utils.GlobalObject.MaxConn)
				conn.Close()
				continue
			}

			// 当处理新链接的业务方法 和 conn 进行绑定，得到我们的链接模块
			dealConn := NewConnection(s, conn, cid, s.MsgHandler)
			cid++

			// 启动当前的业务链接处理
			go dealConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	// 将一些服务器的资源、状态或者一些已经开辟的连接信息 进行停止或回收
	fmt.Println("【STOP】Paguma server name ", s.Name)
	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	// 启动server的服务功能
	s.Start()

	// TODO 做一些启动服务之后的额外业务

	// 阻塞状态
	select {}
}

func (s *Server) GetConnMgr() pgiface.IConnManager {
	return s.ConnMgr
}

// AddRouter 路由功能：给当前的服务注册一个路由方法，供客户端的链接处理使用
func (s *Server) AddRouter(msgID uint32, router pgiface.IRouter) {
	s.MsgHandler.AddRouter(msgID, router)
	fmt.Println("Add router succeed!")
}

// NewServer 初始化Server模块的方法
func NewServer() pgiface.IServer {
	s := &Server{
		Name:       utils.GlobalObject.Name,
		IPVersion:  "tcp4",
		IP:         utils.GlobalObject.Host,
		Port:       utils.GlobalObject.TCPPort,
		MsgHandler: NewMsgHandle(),
		ConnMgr:    NewConnManager(),
	}
	return s
}

//SetOnConnStart 设置该Server的连接创建时Hook函数
func (s *Server) SetOnConnStart(hookFunc func(pgiface.IConnection)) {
	s.OnConnStart = hookFunc
}

//SetOnConnStop 设置该Server的连接断开时的Hook函数
func (s *Server) SetOnConnStop(hookFunc func(pgiface.IConnection)) {
	s.OnConnStop = hookFunc
}

//CallOnConnStart 调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn pgiface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---> CallOnConnStart....")
		s.OnConnStart(conn)
	}
}

//CallOnConnStop 调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn pgiface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("---> CallOnConnStop....")
		s.OnConnStop(conn)
	}
}
