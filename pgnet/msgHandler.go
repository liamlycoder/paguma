package pgnet

import (
	"fmt"
	"paguma/pgiface"
	"strconv"
)

/*
消息处理模块的实现
 */

type MsgHandler struct {
	// 存放每一个MsgID所对应的处理方法
	Apis map[uint32] pgiface.IRouter
}

// NewMsgHandle 创建MsgHandle的方法
func NewMsgHandle() *MsgHandler {
	return &MsgHandler{Apis: make(map[uint32]pgiface.IRouter)}
}

func (m *MsgHandler)DoMsgHandler(request pgiface.IRequest) {
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
func (m *MsgHandler)AddRouter(msgID uint32, router pgiface.IRouter) {
	// 1. 判断当前msg绑定的api处理方法是否已经存在
	if _, ok := m.Apis[msgID]; ok {
		// 已经注册了
		panic("repeat api, msgID: " + strconv.Itoa(int(msgID)))
	}
	// 2. 添加msg与api的绑定关系
	m.Apis[msgID] = router
	fmt.Println("add api to msgID=", msgID, "succeed")
}
