package server

import (
	"apcchis.com/apcc/core"
	"apcchis.com/apcc/handler"

	"apcchis.com/apcc/common"

	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
)

// Version命令处理器
func handleVersion(request []byte) {
	var buff bytes.Buffer

	var payload Version

	dataBytes := request[common.COMMAND_LENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {

		log.Panic(err)
	}
	// 提取最大区块高度作比较
	bestHeight := core.NewestBlock.Height
	foreignerBestHeight := payload.BestHeight

	if bestHeight > foreignerBestHeight {

		// 向请求节点回复自身Version信息
		sendVersion(payload.AddrFrom)
	} else if bestHeight < foreignerBestHeight {

		// 向请求节点要信息
		sendGetBlocks(payload.AddrFrom)
	}

	// 添加到已知节点中
	if !nodeIsKnown(payload.AddrFrom) {

		common.KnowedNodes = append(common.KnowedNodes, payload.AddrFrom)
	}
}

func handleAddr(request []byte) {

}

func handleGetblocks(request []byte) {

	var buff bytes.Buffer
	var payload GetBlocks

	dataBytes := request[common.COMMAND_LENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	spew.Dump("handleGetblocks:", payload)
	blocks := handler.GetBlockHashes()

	sendInv(payload.AddrFrom, common.BLOCK_TYPE, blocks)
}

func handleGetData(request []byte) {

	var buff bytes.Buffer
	var payload GetData

	dataBytes := request[common.COMMAND_LENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {

		log.Panic(err)
	}

	if payload.Type == common.BLOCK_TYPE {

		sendBlock(payload.AddrFrom, handler.GetBlock(payload.Hash))
	}

	if payload.Type == common.TX_TYPE {

		// 取出交易

		for _, tx := range core.MemTxPool {
			if bytes.Compare(tx.TxHash, payload.Hash) == 0 {
				SendTx(payload.AddrFrom, &tx)
			}
		}

	}
}

func handleBlock(request []byte) {

	//fmt.Println("handleblock:\n")
	//blc.Printchain()

	var buff bytes.Buffer
	var payload BlockData

	dataBytes := request[common.COMMAND_LENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {

		log.Panic(err)
	}

	block := core.DeSerializeBlock(payload.BlockBytes)
	if block == nil {

		fmt.Printf("Block nil")
	}

	err = handler.AddBlock(block)
	if err != nil {

		log.Panic(err)
	}
	fmt.Printf("add block %x succ.\n", block.Hash)
	//blc.Printchain()

	if len(common.UnslovedHashes) > 0 {

		sendGetData(payload.AddrFrom, common.BLOCK_TYPE, common.UnslovedHashes[0])
		common.UnslovedHashes = common.UnslovedHashes[1:]
	} else {

		//blc.Printchain()
		core.ResetUTXOSet()
	}
}

func handleTx(request []byte) {

	var buff bytes.Buffer
	var payload TxData

	dataBytes := request[common.COMMAND_LENGTH:]
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {

		log.Panic(err)
	}

	tx := core.DeserializeTransaction(payload.TransactionBytes)

	// 自身为主节点，需要将交易转发给矿工节点
	if common.NodeAddr == common.KnowedNodes[0] {

		for _, node := range common.KnowedNodes {

			if node != common.NodeAddr && node != payload.AddFrom {

				sendInv(node, common.TX_TYPE, [][]byte{tx.TxHash})
			}
		}
	} else {
		core.MemTxPool = append(core.MemTxPool, tx)
	}
}

func handleInv(request []byte) {

	var buff bytes.Buffer
	var payload Inv

	dataBytes := request[common.COMMAND_LENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	spew.Dump("handleInv:", payload)

	// Ivn 3000 block hashes [][]
	if payload.Type == common.BLOCK_TYPE {

		fmt.Println(payload.Items)

		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, common.BLOCK_TYPE, blockHash)

		if len(payload.Items) >= 1 {

			common.UnslovedHashes = payload.Items[1:]
		}
	}

	if payload.Type == common.TX_TYPE {

		TxHash := payload.Items[0]

		// 添加到交易池

		var flag bool
		for _, tx := range core.MemTxPool {
			if bytes.Compare(tx.TxHash, TxHash) == 0 {
				flag = true
			}
		}
		if flag {
			sendGetData(payload.AddrFrom, common.TX_TYPE, TxHash)
		}
	}
}
