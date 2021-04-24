package pgnet

type Message struct {
	Id      uint32 // 消息的ID
	DataLen uint32 // 消息的长度
	Data    []byte // 消息的内容
}

// NewMsgPack 提供一个创建Message的方法
func NewMsgPack(id uint32, data []byte) *Message {
	return &Message{
		Id:      id,
		DataLen: uint32(len(data)),
		Data:    data,
	}
}

// GetMsgId 获取消息的ID
func (m *Message)GetMsgId() uint32 {
	return m.Id
}

// GetMsgLen 获取消息的长度
func (m *Message)GetMsgLen() uint32 {
	return m.DataLen
}

// GetData 获取消息的内容
func (m *Message)GetData() []byte {
	return m.Data
}

// SetMsgId 设置消息的ID
func (m *Message)SetMsgId(id uint32) {
	m.Id = id
}

// SetMsgLen 设置消息的长度
func (m *Message)SetMsgLen(dataLen uint32) {
	m.DataLen = dataLen
}

// SetData 设置消息的内容
func (m *Message)SetData(data []byte) {
	m.Data = data
}
