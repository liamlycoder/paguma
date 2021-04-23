package pgnet

import (
	"fmt"
	"io"
	"net"
	"testing"
)

// 单元测试：测试datapack封包、拆包的问题
func TestDataPack(t *testing.T)  {
	/*
	模拟的服务器
	 */
	// 1. 创建socketTCP
	listener, err := net.Listen("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("server listen err: ", err)
		return
	}

	// 创建一个go去负责从客户端处理业务
	go func() {
		// 2. 从客户端读取数据，拆包处理
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("server accept err: ", err)
				continue
			}
			go func(conn net.Conn) {
				// 处理客户端的请求
				// --- 拆包的过程 ----
				dp := NewDataPack()
				for {
					// 1. 第一次读，把head信息读出来
					headData := make([]byte, dp.GetHeadLen())
					_, err = io.ReadFull(conn, headData)
					if err != nil {
						fmt.Println("read head err: ", err)
						break
					}

					msgHead, err := dp.Unpack(headData)
					if err != nil {
						fmt.Println("server unpack err: ", err)
						return
					}
					if msgHead.GetMsgLen() > 0 {
						// Msg是有数据的，需要进行第二次读取
						// 2. 第二次读，根据head中的dataLen再读取消息内容
						msg := msgHead.(*Message)   // 类型断言
						msg.Data = make([]byte, msg.GetMsgLen())
						_, err := io.ReadFull(conn, msg.Data)
						if err != nil {
							fmt.Println("server unpack err: ", err)
							return
						}

						// 完整的一个消息已经读取完毕
						fmt.Println("--->Recv MsgID: ", msg.Id, ", dataLen = ", msg.DataLen, ", data = ", string(msg.Data))

					}

				}
			}(conn)
		}
	}()


	/*
	模拟客户端
	 */
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("client dial err: ", err)
		return
	}

	// 创建一个封包对象datapack
	dp := NewDataPack()

	// 模拟粘包过程，封装两个msg一同发送

	// 封装第一个msg1包
	msg1 := &Message{
		Id:      1,
		DataLen: 6,
		Data:    []byte{'H', 'e', 'z', 'h', 'a', 'o'},
	}
	sendData1, err := dp.Pack(msg1)
	if err != nil {
		fmt.Println("client pack msg1 err: ", err)
		return
	}

	// 封装第二个msg2包
	msg2 := &Message{
		Id:      2,
		DataLen: 4,
		Data:    []byte{'l', 'u', 'y', 'u'},
	}
	sendData2, err := dp.Pack(msg2)
	if err != nil {
		fmt.Println("client pack msg2 err: ", err)
		return
	}

	// 将两个包粘在一起
	sendData1 = append(sendData1, sendData2...)  // 这里需要打散

	// 一同发送给服务器
	_, _ = conn.Write(sendData1)

	// 客户端阻塞
	select {}
}
