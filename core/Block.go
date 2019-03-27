package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"math/big"
	"time"

	"apcchis.com/apcc/consensus"
	"apcchis.com/apcc/db"
	"github.com/sirupsen/logrus"

	"apcchis.com/apcc/common"
)

type Block struct {
	//1.区块高度
	Height int64
	//2.上一个区块HAsh
	PrevBlockHash []byte
	//3.交易数据
	Txs []*Transaction
	//4.时间戳
	Timestamp int64
	//5.Hash
	Hash  []byte
	Nonce *big.Int
}

var NewestBlock *Block

//1.创建新的区块
func NewBlock(txs []*Transaction, height int64, prevBlockHash []byte) *Block {

	//创建区块
	block := &Block{
		Height:        height,
		PrevBlockHash: prevBlockHash,
		Txs:           txs,
		Timestamp:     time.Now().Unix(),
		Hash:          nil,
		Nonce:         big.NewInt(0),
	}
	//工作量证明预留
	//调用工作量证明返回有效的Hash
	// pow := NewProofOfWork(block)
	pos := consensus.NewProofOfWork()

Calut:
	block.Timestamp = time.Now().Unix()
	// t, _ := time.ParseInLocation("2006-01-02 15:04:05", "2019-03-24 01:00:00", time.Local)
	// block.Timestamp = t.Unix()
	block.Hash = block.GetHash()
	block.Nonce = pos.Run(block.PrevBlockHash, block.Hash)

	if pos.IsValid(block.PrevBlockHash, block.Hash, block.Nonce) {
		logrus.Infof("\r出块成功,高度%d\n", block.Height)
		return block
	}

	hash := sha256.New()
	hash.Write(bytes.Join([][]byte{block.PrevBlockHash, block.Hash}, []byte{}))
	start := new(big.Int).SetBytes(hash.Sum(nil))
	logrus.Errorf("\r出块失败,区块VDF校验失败,p=%s,start=%d,Nonce=%s \n", consensus.P64, start, block.Nonce)
	goto Calut
}

//拼接区块属性，返回字节数组
func (block *Block) GetHash() []byte {

	data := bytes.Join(
		[][]byte{
			block.PrevBlockHash,
			block.HashTransactions(),
			common.IntToHex(block.Timestamp),
			common.IntToHex(int64(common.ComputationSteps)),
			common.IntToHex(int64(block.Height)),
		},
		[]byte{},
	)
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

//单独方法生成创世区块
func CreateGenesisBlock(txs []*Transaction) *Block {

	return NewBlock(
		txs,
		1,
		nil,
	)
}

//区块序列化
func (block *Block) Serialize() []byte {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

//区块反序列化
func DeSerializeBlock(blockBytes []byte) *Block {

	block := new(Block)
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))

	err := decoder.Decode(block)

	if err != nil {

		logrus.Panic(err)
	}

	return block
}

// 需要将Txs转换成[]byte
func (block *Block) HashTransactions() []byte {

	//引入MerkleTree前的交易哈希
	//var TxHashes [][]byte
	//var TxHash [32]byte
	//
	//for _, tx := range block.Txs {
	//
	//	TxHashes = append(TxHashes, tx.TxHash)
	//}
	//TxHash = sha256.Sum256(bytes.Join(TxHashes, []byte{}))
	//
	//return TxHash[:]

	//默克尔树根节点表示交易哈希
	var transactions [][]byte

	for _, tx := range block.Txs {

		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
}

//创币交易判断
func (block *Block) IsGenesisBlock() bool {
	return block.Height == 1 && len(block.PrevBlockHash) == 0
}

func init() {
	logrus.Debugln("init value1")

	logrus.Debugf("value =====   \n\n")
	defer db.Close()
	if value := db.Get(common.BlockTableName, []byte(common.NewestBlockKey)); len(value) > 0 {
		logrus.Debugln("init value", value)
		// spew.Dump(NewestBlock)
		if value = db.Get(common.BlockTableName, value); len(value) > 0 {
			NewestBlock = DeSerializeBlock(value)
			logrus.Infof("最新区块高度=%x", NewestBlock.Height)
			return
		}
		logrus.Errorln("获取最新区块出错")
	}
}
