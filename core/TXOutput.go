package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"

	"apcchis.com/apcc/common"
)

type TXOutput struct {
	//面值
	Coin int64

	//用户名
	Ripemd160Hash []byte //用户名  公钥两次哈希后的值
	//千分之2冻结
	Frozen bool
}
type UTXO struct {
	//来自交易的哈希
	TxHash []byte
	//在该交易VOuts里的下标
	Index int

	//未花费的交易输出
	Output *TXOutput
	//锁定时间 区块生成时间
	LockTime int64
}

type TXOutputs struct {
	UTXOS []*UTXO
}

func NewTXOutput(coin int64, address string, frozen bool) *TXOutput {

	txOutput := &TXOutput{coin, nil, frozen}

	// 设置Ripemd160Hash
	txOutput.Lock(address)

	return txOutput
}

//锁定
func (utxo *UTXO) GetBlacnce() (useableAmount, frozenAmount int64) {
	if utxo.Output.Frozen { //该输出被锁仓
		durtH := time.Now().Sub(time.Unix(utxo.LockTime, 0)).Hours()
		logrus.Debugln("锁定时间", time.Unix(utxo.LockTime, 0).Format("2006-01-02 15:04:05"), "相差", durtH)
		if durtH > 1 { //千分之2释放
			useableAmount = utxo.Output.Coin / 1000 * 2 * int64(durtH/24)
			frozenAmount = utxo.Output.Coin - useableAmount
			logrus.Debugln("useableAmount=", useableAmount, "frozenAmount", frozenAmount)
		} else {
			frozenAmount = utxo.Output.Coin
		}
		return
	}
	useableAmount = utxo.Output.Coin
	return
}

//锁定
func (txOutput *TXOutput) Lock(address string) {

	version_pubKeyHash_checkSumBytes := common.Base58Decode([]byte(address))
	txOutput.Ripemd160Hash = version_pubKeyHash_checkSumBytes[1 : len(version_pubKeyHash_checkSumBytes)-4]
}

//解锁
func (txOutput *TXOutput) UnLockScriptPubKeyWithAddress(address string) bool {

	version_pubKeyHash_checkSumBytes := common.Base58Decode([]byte(address))
	ripemd160Hash := version_pubKeyHash_checkSumBytes[1 : len(version_pubKeyHash_checkSumBytes)-4]

	fmt.Printf(" address=%s  \n %x\n %x\n", address, txOutput.Ripemd160Hash, ripemd160Hash)
	return bytes.Compare(txOutput.Ripemd160Hash, ripemd160Hash) == 0
}

// 反序列化
func DeserializeTXOutputs(txOutputsBytes []byte) *TXOutputs {

	var txOutputs TXOutputs
	decoder := gob.NewDecoder(bytes.NewReader(txOutputsBytes))
	err := decoder.Decode(&txOutputs)
	if err != nil {
		log.Panic(err)
	}

	return &txOutputs
}

// 序列化成字节数组
func (txOutputs *TXOutputs) Serialize() []byte {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(txOutputs)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}
