package server

type Version struct {
	// 版本
	Version int64
	// 当前节点区块的高度
	BestHeight int64
	//当前节点的地址
	AddrFrom string
}

// 用于请求区块或交易
type GetData struct {
	// 节点地址
	AddrFrom string
	// 请求类型  是block还是tx
	Type string
	// 区块哈希或交易哈希
	Hash []byte
}

// 表示向节点请求一个块哈希的表，该请求会返回所有块的哈希
type GetBlocks struct {
	//请求节点地址
	AddrFrom string
}

// 用于节点间发送一个区块
type BlockData struct {
	// 节点地址
	AddrFrom string
	// 序列化区块
	BlockBytes []byte
}

// 同步中传递的交易类型
type TxData struct {
	// 节点地址
	AddFrom string
	// 交易
	TransactionBytes []byte
}

// 向其他节点展示自己拥有的区块和交易
type Inv struct {
	// 自己的地址
	AddrFrom string
	// 类型 block tx
	Type string
	// hash二维数组
	Items [][]byte
}
