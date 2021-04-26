package pgnet

import (
	"fmt"
	"paguma/pgiface"
	"paguma/utils"
	"strconv"
)

/*
消息处理模块的实现
*/

type MsgHandler struct {
	// 存放每一个MsgID所对应的处理方法
	Apis map[uint32]pgiface.IRouter
	// 业务工作Worker池的数量
	WorkerPoolSize uint32
	// 负责Worker取任务的消息队列
	TaskQueue []chan pgiface.IRequest
}

// NewMsgHandle 创建MsgHandle的方法
func NewMsgHandle() *MsgHandler {
	return &MsgHandler{
		Apis:           make(map[uint32]pgiface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		TaskQueue:      make([]chan pgiface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}

func (m *MsgHandler) DoMsgHandler(request pgiface.IRequest) {
	// 1. 从request中找到msgID，并从apis中找到对应的处理api
	handler, ok := m.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgID = ", request.GetMsgID(), " is not found. Please register firstly")
	}
	// 2. 调度对应的router业务
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}
func (m *MsgHandler) AddRouter(msgID uint32, router pgiface.IRouter) {
	// 1. 判断当前msg绑定的api处理方法是否已经存在
	if _, ok := m.Apis[msgID]; ok {
		// 已经注册了
		panic("repeat api, msgID: " + strconv.Itoa(int(msgID)))
	}
	// 2. 添加msg与api的绑定关系
	m.Apis[msgID] = router
	fmt.Println("add api to msgID=", msgID, "succeed")
}

// StartWorkerPool 启动一个Worker工作池 (开启工作池的动作只能发生一次，因为一个paguma框架只能有一个工作池）
func (m *MsgHandler) StartWorkerPool() {
	// 根据workerPoolSize来分别开启worker，每个worker用一个goroutine来承载
	for i := 0; i < int(m.WorkerPoolSize); i++ {
		// 一个worker被启动
		// 当前的worker对应的channel消息队列，开辟空间，第0个worker就用第0个channel...
		m.TaskQueue[i] = make(chan pgiface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		// 启动当前的worker，阻塞等待消息从channel传递进来
		go m.StartOneWorker(i, m.TaskQueue[i])

	}
}

// StartOneWorker 启动一个Worker工作流程
func (m *MsgHandler) StartOneWorker(workerID int, taskQueue chan pgiface.IRequest) {
	fmt.Println("WorkerID = ", workerID, " is started...")
	// 不断阻塞等待对应消息队列的消息
	for {
		select {
		// 如果有消息过来，出列的就是一个客户端的request，执行当前Request所绑定的业务
		case request := <-taskQueue:
			m.DoMsgHandler(request)
		}
	}
}

func (m *MsgHandler)SendMsgToTaskQueue(request pgiface.IRequest)  {
	// 1. 将消息平均分配给不同的worker
	// 根据客户端建立的connID来分配
	workerID := request.GetConnection().GetConnID() % m.WorkerPoolSize
	fmt.Println("add connID = ", request.GetConnection().GetConnID(), " request MsgID = ", request.GetMsgID(), " to WorkerID = ", workerID)
	// 2. 将消息发送给对应的worker的TaskQueue即可
	m.TaskQueue[workerID] <- request
}
