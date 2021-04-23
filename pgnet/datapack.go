package pgnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"paguma/pgiface"
	"paguma/utils"
)

/*
封包、拆包的具体模块
 */

type DataPack struct {

}

// NewDataPack 封包、拆包实例的初始化方法
func NewDataPack() *DataPack {
	return &DataPack{}
}

// GetHeadLen 获取包的头的长度
func (d *DataPack)GetHeadLen() uint32 {
	// DataLen uint32(4字节) + ID uint32(4字节) = 8字节
	return 8
}

// Pack 封包
func (d *DataPack)Pack(msg pgiface.IMessage) ([]byte, error) {
	// 创建一个存放bytes字节的缓冲
	dataBuf := bytes.NewBuffer([]byte{})

	// 包的格式：|dataLen|MsgID|data|

	// 将dataLen写进dataBuf中
	if err := binary.Write(dataBuf, binary.LittleEndian, msg.GetMsgLen()); err != nil {
		return nil, err
	}

	// 将MsgId写进dataBuf中
	if err := binary.Write(dataBuf, binary.LittleEndian, msg.GetMsgId()); err != nil {
		return nil, err
	}

	// 将data数据写进dataBuf中
	if err := binary.Write(dataBuf, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}

	return dataBuf.Bytes(), nil
}

// Unpack 拆包。将包的Head读出来之后，再根据head的信息里的data长度，再进行一次读
func (d *DataPack)Unpack(binaryData[]byte) (pgiface.IMessage, error) {
	// 创建一个从输入二进制数据的ioReader
	dataBuf := bytes.NewReader(binaryData)

	// 只解压head信息，得到dataLen和MsgID
	msg := &Message{}


	// 读dataLen
	if err := binary.Read(dataBuf, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	// 读MsgID
	if err := binary.Read(dataBuf, binary.LittleEndian, &msg.Id); err != nil {
		return nil, err
	}

	// 判断dataLen是否已经超出了我们允许的最大包长度
	if utils.GlobalObject.MaxPacketSize > 0 && msg.DataLen > utils.GlobalObject.MaxPacketSize {
		return nil, errors.New("too large message data")
	}

	return msg, nil
}

