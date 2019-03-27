package server

import (
	"bufio"
	"fmt"
	"sync"

	"apcchis.com/apcc/common"
)

var mutex = &sync.Mutex{}

func readData() {
	fmt.Println("readData")
	for {

		buf := make([]byte, 1024) // 这个1024可以根据你的消息长度来设置
		n, err := stm.Read(buf)   // n为一次Read实际得到的消息长度

		if err != nil {
			fmt.Println("read error:", err)
			break
		}
		fmt.Println("bts", string(buf[:n]))
		go handleConnection(buf[:n])
	}

}

func writeData(data []byte) {
	rw := bufio.NewReadWriter(bufio.NewReader(stm), bufio.NewWriter(stm))
	mutex.Lock()
	fmt.Printf("send--Data: %x", data)
	rw.WriteString(fmt.Sprintf("%s\n", string(data)))
	rw.Flush()
	mutex.Unlock()

}

// 客户端命令处理器
func handleConnection(request []byte) {
	fmt.Println("handleConnection:\n", string(request))

	//blc.Printchain()

	// 读取客户端发送过来的所有的数据

	fmt.Printf("\nReceive a Message:%s\n", request[:common.COMMAND_LENGTH])
	fmt.Printf("\ndata:%s\n", string(request))

	command := common.BytesToCommand(request[:common.COMMAND_LENGTH])

	switch command {

	case common.COMMAND_VERSION:
		handleVersion(request)

	case common.COMMAND_ADDR:
		handleAddr(request)

	case common.COMMAND_BLOCK:
		handleBlock(request)

	case common.COMMAND_GETBLOCKS:
		handleGetblocks(request)

	case common.COMMAND_GETDATA:
		handleGetData(request)

	case common.COMMAND_INV:
		handleInv(request)

	case common.COMMAND_TX:
		handleTx(request)

	default:
		fmt.Println("Unknown command!")
	}

}

// 节点是否在已知节点中
func nodeIsKnown(addr string) bool {

	for _, node := range common.KnowedNodes {

		if node == addr {

			return true
		}
	}

	return false
}
