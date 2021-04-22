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
	// 当前的Server添加一个router, server注册的链接对应的处理业务
	Router pgiface.IRouter
}



func (s *Server)Start()  {
	fmt.Printf("[Paguma] Server Name: %s, listener IP at : %s, Port : %d is starting...\n", utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TCPPort )
	fmt.Printf("[Paguma] Version %s, MaxConn: %d, MaxPacketSize: %d\n", utils.GlobalObject.Version, utils.GlobalObject.MaxConn, utils.GlobalObject.MaxPacketSize)
	go func() {
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

			// 当处理新链接的业务方法 和 conn 进行绑定，得到我们的链接模块
			dealConn := NewConnection(conn, cid, s.Router)
			cid++

			// 启动当前的业务链接处理
			go dealConn.Start()
		}
	}()
}

func (s *Server)Stop()  {
	// TODO 将一些服务器的资源、状态或者一些已经开辟的连接信息 进行停止或回收
}

func (s *Server)Serve()  {
	// 启动server的服务功能
	s.Start()

	// TODO 做一些启动服务之后的额外业务

	// 阻塞状态
	select {}
}

// AddRouter 路由功能：给当前的服务注册一个路由方法，供客户端的链接处理使用
func (s *Server)AddRouter(router pgiface.IRouter) {
	s.Router = router
	fmt.Println("Add router succeed!")
}

// NewServer 初始化Server模块的方法
func NewServer(name string) pgiface.IServer {
	s := &Server{
		Name:      utils.GlobalObject.Name,
		IPVersion: "tcp4",
		IP:        utils.GlobalObject.Host,
		Port:      utils.GlobalObject.TCPPort,
		Router:    nil,
	}
	return s
}



