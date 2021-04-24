package main

import (
	"fmt"
	"io"
	"net"
	"paguma/pgnet"
	"time"
)

// 模拟客户端

func main() {
	fmt.Println("Client0 start...")
	time.Sleep(time.Second)
	// 1. 直接连接远程服务器，得到一个conn连接
	conn, err := net.Dial("tcp", "127.0.0.1:8997")
	if err != nil {
		fmt.Println("client dial err: ", err)
		return
	}

	for {
		// 发送封包的message消息
		dp := pgnet.NewDataPack()
		binaryMsg, err := dp.Pack(pgnet.NewMsgPack(0, []byte("Hello, H_zZZ")))
		if err != nil {
			fmt.Println("pack error: ", err)
			return
		}
		if _, err := conn.Write(binaryMsg); err != nil {
			fmt.Println("write error: ", err)
			return
		}

		// 服务器就应该给我们回复一个message数据：MsgID: 1  ping...ping...ping
		// 1. 先读取流中的head部分，得到ID和dataLen
		binaryHead := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(conn, binaryHead); err != nil {
			fmt.Println("read head error: ", err)
			break
		}
		// 将二进制的head拆包到msg结构中
		msgHead, err := dp.Unpack(binaryHead)
		if err != nil {
			fmt.Println("client unpack msgHead error: ", err)
			break
		}
		if msgHead.GetMsgLen() > 0 {
			// 2. 再根据DataLen进行第二次读取，将data读出来
			msg := msgHead.(*pgnet.Message)
			msg.Data = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(conn, msg.Data); err != nil {
				fmt.Println("read msg data error: ", err)
				return
			}
			fmt.Println("--->receive server msg: ID = ", msg.GetMsgId(), ", length = ", msg.GetMsgLen(), ", data = ", string(msg.GetData()))
		}

		// cpu阻塞
		time.Sleep(time.Second)
	}
}

