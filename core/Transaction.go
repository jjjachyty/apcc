package core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/boltdb/bolt"

	btcutil "github.com/FactomProject/btcutilecc"
	"github.com/sirupsen/logrus"

	"apcchis.com/apcc/db"
	"apcchis.com/apcc/wallet"

	"apcchis.com/apcc/common"
) // 交易内存池
var MemTxPool = make([]Transaction, 0)

type Transaction struct {
	//1.交易哈希值
	TxHash []byte
	//2.输入
	Vins []*TXInput
	//3.输出
	Vouts []*TXOutput
}

//交易序列化
func (tx *Transaction) Serialize() []byte {

	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	//tx.PrintTx()
	//fmt.Printf("\n%x\n", encoded.Bytes())

	return encoded.Bytes()
}

//将交易信息转换为字节数组
func (tx *Transaction) HashTransactions() {

	//交易信息序列化
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(tx)
	if err != nil {

		log.Panic(err)
	}

	//是创币交易的哈希不同
	timeSpBytes := common.IntToHex(time.Now().Unix())
	//设置hash
	TxHash := sha256.Sum256(bytes.Join([][]byte{timeSpBytes, result.Bytes()}, []byte{}))
	tx.TxHash = TxHash[:]
}

//创币交易判断
func (tx *Transaction) IsCoinbaseTransaction() bool {

	return len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Vout == -1
}

// 拷贝一份新的Transaction用于签名,包含所有的输入输出，但TXInput.Signature 和 TXIput.PubKey 被设置为 nil                                 T
func (tx *Transaction) TrimmedCopy() Transaction {

	var inputs []*TXInput
	var outputs []*TXOutput

	for _, vin := range tx.Vins {

		inputs = append(inputs, &TXInput{vin.TxHash, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vouts {

		outputs = append(outputs, &TXOutput{vout.Coin, vout.Ripemd160Hash, vout.Frozen})
	}

	txCopy := Transaction{tx.TxHash, inputs, outputs}

	//fmt.Printf("\ntx:\n%x\ncopy:\n%x", tx.TxHash, txCopy.TxHash)

	return txCopy
}

//1.创建创世区块
func CreateBlockchainWithGensisBlock() {
	if NewestBlock == nil { //没有区块
		//创币交易
		txCoinbase := NewCoinbaseTransaction(wallet.MyWallet.ExternalAddress[0].Value, common.CoinBase)

		//创世区块
		NewestBlock = CreateGenesisBlock([]*Transaction{txCoinbase})

		if err := db.UpdateFunc(func(tx *bolt.Tx) (err error) {
			var b *bolt.Bucket
			if b, err = tx.CreateBucketIfNotExists([]byte(common.BlockTableName)); err == nil {
				if err = b.Put(NewestBlock.Hash, NewestBlock.Serialize()); err == nil {
					if err = b.Put([]byte(common.NewestBlockKey), NewestBlock.Hash); err == nil {
						logrus.Debugf("创建创世区块成功 HASH=%x \n", NewestBlock.Hash)
						return
					}
				}
				logrus.Fatal(err)
				return
			}

			return nil
		}); err != nil {

			logrus.Errorln("创建创世块失败")
			return
		}
		ResetUTXOSet()
	}

}

//1.创世币交易
func NewCoinbaseTransaction(address string, coin int64) *Transaction {
	logrus.Debugf("为地址%s 创建交易\n", address)
	//输入  由于创世区块其实没有输入，所以交易哈希传空，TXOutput索引传-1，签名随你
	txInput := &TXInput{[]byte{}, -1, []byte{}, []byte{}}
	//输出  产生一笔奖励给挖矿者
	txOutput := NewTXOutput(coin, address, false)
	txCoinbase := &Transaction{
		[]byte{},
		[]*TXInput{txInput},
		[]*TXOutput{txOutput},
	}

	txCoinbase.HashTransactions()
	logrus.Debugf("TXHash =%x \n", txCoinbase.TxHash)
	return txCoinbase
}

//NewTransaction 普通交易
func NewTransaction(from string, changeType int, to string, amount int64, free int64, frozen bool, txs []*Transaction) (*Transaction, error) {
	//输入输出
	var txInputs []*TXInput
	var txOutputs []*TXOutput
	if free == 0 {
		free = common.OneCoin
	}
	userableAmount, frozenAmount, spendableUTXODic := FindSpendableUTXOs(from, amount+free, txs)
	logrus.Debugf("账户【%s】找到【%d】转账给账户【%s】【%d】APCC 剩余待释放金额【%d】", from, userableAmount, to, amount, frozenAmount)
	if userableAmount >= amount {
		var pk = wallet.MyWallet.GetPrivKey(from)
		var pubk = common.PriveKeyToPublicKey(pk.D.Bytes())

		for TxHash, indexArr := range spendableUTXODic {
			fmt.Printf("引用未花费的交易 %s\n", TxHash)
			//字符串转换为[]byte
			TxHashBytes, _ := hex.DecodeString(TxHash)
			for _, index := range indexArr {

				//交易输入
				txInput := &TXInput{
					TxHashBytes,
					index,
					nil,
					pubk,
				}

				txInputs = append(txInputs, txInput)
			}
		}
		//可用余额有多余的
		if change := userableAmount - amount - free; change > 0 {

			//找零
			var changeAddr = from
			if changeType == 1 { //内部地址找零
				changeAddr = wallet.MyWallet.GetNewAddress(1).Value
			}
			logrus.Debugf("地址【%s】找零【%d】", changeAddr, -change)
			txOutput := NewTXOutput(change, changeAddr, false)
			txOutputs = append(txOutputs, txOutput)
		}
		//冻结金额返回给原始账户继续冻结
		if frozenAmount > 0 {
			//转账
			txOutput := NewTXOutput(frozenAmount, from, true)
			txOutputs = append(txOutputs, txOutput)
		}
		//转账给别人
		txOutput := NewTXOutput(int64(amount), to, frozen)
		txOutputs = append(txOutputs, txOutput)

		//交易构造
		tx := &Transaction{
			[]byte{},
			txInputs,
			txOutputs,
		}

		tx.HashTransactions()

		//进行签名
		tx.SignTransaction(pk, txs)
		logrus.Debugf("返回交易%x\n", tx.TxHash)
		return tx, nil
	}

	logrus.Errorf("地址 %s 余额%d ,无法转账%d 待释放余额%d", from, userableAmount, amount, frozenAmount)

	return nil, errors.New("余额不足")
}

//交易签名
func (tx *Transaction) SignTransaction(privKey *ecdsa.PrivateKey, txs []*Transaction) {

	if tx.IsCoinbaseTransaction() {

		return
	}
	var prevTX Transaction
	var err error
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vins {

		//找到当前交易输入引用的所有交易
		//fmt.Printf("txHas0:%x\n", vin.TxHash)
		prevTX, err = FindTransaction(vin.TxHash, txs)
		if err != nil {
			logrus.Errorln(err)
		}

		prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX
	}
	tx.Sign(privKey, prevTXs)
}

//数字签名
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey, prevTxs map[string]Transaction) {

	//判断当前交易是否为创币交易，coinbase交易因为没有实际输入，所以没有被签名
	if tx.IsCoinbaseTransaction() {

		return
	}

	//将会被签署的是修剪后的交易副本

	txCopy := tx.TrimmedCopy()

	//遍历交易的每一个输入
	for inID, vin := range txCopy.Vins {

		//交易输入引用的上一笔交易
		prevTx := prevTxs[hex.EncodeToString(vin.TxHash)]
		//Signature 被设置为 nil
		txCopy.Vins[inID].Signature = nil
		fmt.Println("ind", inID, "vin.Vout", vin.Vout, "len", prevTx)
		//PubKey 被设置为所引用输出的PubKeyHash
		txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash

		// 签名代码
		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, privateKey, []byte(dataToSign))
		if err != nil {
			logrus.Errorf("交易签名失败 %v", err)
		}
		//一个ECDSA签名就是一对数字，我们对这对数字连接起来就是signature
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vins[inID].Signature = signature

		txCopy.Vins[inID].PublicKey = nil

	}

	// fmt.Println("签名完成后开始校验", tx.Verify(prevTxs))
}

// 交易验签
func (tx *Transaction) VerifyTransaction(txs []*Transaction) bool {
	if tx.IsCoinbaseTransaction() {

		return true
	}
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vins {

		prevTX, err := FindTransaction(vin.TxHash, txs)

		if err != nil {
			logrus.Errorf("输入[%x]未找到未花费的交易", vin.TxHash)
			return false
		}
		logrus.Debugf("找到TX %v", prevTX)
		prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX
	}

	return tx.Verify(prevTXs)
}

//获取某个交易
func FindTransaction(txHash []byte, txs []*Transaction) (tx Transaction, err error) {
	key := NewestBlock.Hash

	for _, tx := range txs {

		logrus.Debugf("查找交易%x   %x\n\n", txHash, tx.TxHash)
		if bytes.Compare(tx.TxHash, txHash) == 0 {

			return *tx, nil

		}
	}
	logrus.Debugf("从区块链上查找交易%x", txHash)
	defer db.Close()
	for {
		logrus.Debugf("Get key %x\n\n", key)

		blockBytes := db.Get(common.BlockTableName, key)
		if len(blockBytes) > 0 && len(tx.TxHash) == 0 {

			block := DeSerializeBlock(blockBytes)
			for i, txb := range block.Txs {
				logrus.Debugf("区块-高度[%d] 第%d个 txHash %x TxHash[%x]", block.Height, i, txHash, txb.TxHash)
				if bytes.Compare(txb.TxHash, txHash) == 0 {
					logrus.Debugf("找到了 %v", txb.Vouts)
					tx = *txb
					break
				}
			}
			if block.IsGenesisBlock() || len(tx.TxHash) > 0 {
				logrus.Debugln("结束查找区块高度", block.Height)
				break
			}
			fmt.Printf("blockHASH== %x   PrevBlockHash==  %x \n", block.Hash, block.PrevBlockHash)
			key = block.PrevBlockHash
			continue
		}
		logrus.Debugf("结束查找key=%x byt=%x,tx=%v\n", key, blockBytes, tx)
		break
	}

	return
}

// 验签
func (tx *Transaction) Verify(prevTxs map[string]Transaction) bool {

	if tx.IsCoinbaseTransaction() {

		return true
	}

	for _, vin := range tx.Vins {

		if prevTxs[hex.EncodeToString(vin.TxHash)].TxHash == nil {

			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	//fmt.Println("Verify:")
	txCopy := tx.TrimmedCopy()

	//用于椭圆曲线算法生成秘钥对
	curve := btcutil.Secp256k1()

	// 遍历输入，验证签名
	for inID, vin := range tx.Vins {

		//fmt.Println("Verify:")
		// 这个部分跟Sign方法一样,因为在验证阶段,我们需要的是与签名相同的数据。
		prevTx := prevTxs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[inID].Signature = nil
		txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash

		//txCopy.PrintTx()
		//txCopy.TxHash =  txCopy.Hash()

		//fmt.Println("Verify:")
		//tx.PrintTx()
		//fmt.Println("txCopy:")
		//txCopy.PrintTx()

		// 私钥
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		// 公钥
		puk := common.ExpandPublicKey(vin.PublicKey)

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		// 使用从输入提取的公钥创建ecdsa.PublicKey
		rawPubKey := ecdsa.PublicKey{curve, puk.X, puk.Y}

		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false
		}
		txCopy.Vins[inID].PublicKey = nil
	}

	return true
}

func DeserializeTransaction(data []byte) Transaction {

	var tx Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&tx)
	if err != nil {

		log.Panic(err)
	}

	return tx
}
