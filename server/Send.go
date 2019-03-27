package server

import (
	"apcchis.com/apcc/common"
	"apcchis.com/apcc/core"
)

//COMMAND_VERSION
func sendVersion(toAddress string) {

	bestHeight := core.NewestBlock.Height
	payload := common.GobEncode(Version{common.NODE_VERSION, bestHeight, common.NodeAddr})

	request := append(common.CommandToBytes(common.COMMAND_VERSION), payload...)
	sendData(toAddress, request)
}

//COMMAND_GETBLOCKS
func sendGetBlocks(toAddress string) {

	payload := common.GobEncode(GetBlocks{common.NodeAddr})

	request := append(common.CommandToBytes(common.COMMAND_GETBLOCKS), payload...)

	sendData(toAddress, request)

}

// 主节点将自己的所有的区块hash发送给钱包节点
//COMMAND_BLOCK
//
func sendInv(toAddress string, kind string, hashes [][]byte) {

	payload := common.GobEncode(Inv{common.NodeAddr, kind, hashes})

	request := append(common.CommandToBytes(common.COMMAND_INV), payload...)

	sendData(toAddress, request)

}

func sendGetData(toAddress string, kind string, blockHash []byte) {

	payload := common.GobEncode(GetData{common.NodeAddr, kind, blockHash})

	request := append(common.CommandToBytes(common.COMMAND_GETDATA), payload...)

	sendData(toAddress, request)
}

func sendBlock(toAddress string, blockBytes []byte) {

	payload := common.GobEncode(BlockData{common.NodeAddr, blockBytes})

	request := append(common.CommandToBytes(common.COMMAND_BLOCK), payload...)

	sendData(toAddress, request)
}

func SendTx(toAddress string, tx *core.Transaction) {

	data := TxData{common.NodeAddr, tx.Serialize()}
	payload := common.GobEncode(data)
	request := append(common.CommandToBytes(common.COMMAND_TX), payload...)

	sendData(toAddress, request)

}

// 客户端向服务器发送数据
func sendData(to string, data []byte) {
	writeData(data)
}
