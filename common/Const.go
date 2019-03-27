package common

import (
	"os"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
)

//用于生成地址的校验和位数
const (
	AddressChecksumLen = 4

	//Apcc coin 代码
	Purpose   = 44
	ConinType = 334

	//相关数据库属性

	//最新区块bucket
	NewestTableName = "NBKBUT"
	//最新区块的hash码key
	NewestBlockKey = "NBK"

	DBName         = "apcc.db"
	BlockTableName = "apccbc"

	//存储未花费交易输出的数据库表
	UTXOTableName = "UTXOAPCC"

	//计算次数
	ComputationSteps = 200000

	//初始大小27亿后面8位小数
	CoinBase = 270000000000000000
	OneCoin  = 100000000
	//MiningCost 矿工挖矿消耗
	MiningCost = 5

	Ns         = "apcc"
	ProtocolID = "apcc/1.0.0"

	//发送消息的前12个字节指定了命令名(version)
	COMMAND_LENGTH = 12
	NODE_VERSION   = 1

	// 命令
	COMMAND_VERSION   = "version"
	COMMAND_ADDR      = "addr"
	COMMAND_BLOCK     = "block"
	COMMAND_INV       = "inv"
	COMMAND_GETBLOCKS = "getblocks"
	COMMAND_GETDATA   = "getdata"
	COMMAND_TX        = "tx"

	// 类型
	BLOCK_TYPE = "block"
	TX_TYPE    = "tx"
)

var (
	WalletPassWd = ""
	//P2P监听端口
	ListenPort = 3333
	// MiningAwardAddress 矿工挖矿奖励
	MiningAwardAddress = ""

	//ServerNodeAddrs 全网主节点
	ServerNodeAddrs []string
	KnowedNodes     []string
	//NodeAddr节点地址
	NodeAddr string

	// PublicKeyCompressedLength is the byte count of a compressed public key
	PublicKeyCompressedLength = 33

	// 存储拥有最新链的未处理的区块hash值
	UnslovedHashes [][]byte
)

func init() {

	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	logrus.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
	logrus.SetLevel(logrus.DebugLevel)
}
