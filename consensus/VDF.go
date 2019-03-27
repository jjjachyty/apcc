package consensus

import (
	"bytes"
	"crypto/sha256"
	"time"

	"github.com/sirupsen/logrus"

	"apcchis.com/apcc/common"

	"apcchis.com/apcc/consensus/vdf"

	"math/big"
)

// var p1024 string = "26665316952145251691159678627219217222885850903741016853585447718947343212288750750268012668712469908106258613976547496870890438504017231007766799519535785905104605162203896873810538315838185502276890025696087480171103337359532995917850779890238106057070346163136946293278160601772800244012833993583077700483"

// var p512 string = "1428747867218506432894623188342974573745986827958686951828141301796511703204477877094047850395093527438571991358833787830431256534283107665764428020239091"
// var p256 string = "60464814417085833675395020742168312237934553084050601624605007846337253615407"
// var p128 string = "271387921886905605025992265577018698667"
var P64 string = "9853393445385562019"

type ProofOfStake struct {
	//求工作量的block

}

func NewProofOfWork() *ProofOfStake {
	return &ProofOfStake{}
}

//判断当前区块是否有效
func (pos *ProofOfStake) IsValid(prevBlockHash, currentBlockHash []byte, nonce *big.Int) bool {

	hash := sha256.New()
	hash.Write(bytes.Join([][]byte{prevBlockHash, currentBlockHash}, []byte{}))
	start := new(big.Int).SetBytes(hash.Sum(nil))

	// start := new(big.Int).SetBytes(bytes.Join([][]byte{prevBlockHash, currentBlockHash}, []byte{}))

	p, _ := new(big.Int).SetString(P64, 0)

	return vdf.Verify(common.ComputationSteps, start, nonce, p)
}

//运行工作量证明
func (pos *ProofOfStake) Run(prevBlockHash, currentBlockHash []byte) *big.Int {

	logrus.Infoln("正在挖矿...")
	startTime := time.Now()
	hash := sha256.New()
	hash.Write(bytes.Join([][]byte{prevBlockHash, currentBlockHash}, []byte{}))
	start := new(big.Int).SetBytes(hash.Sum(nil))

	p, _ := new(big.Int).SetString(P64, 0)
	result := vdf.Modsqrt_op(common.ComputationSteps, start, p)

	logrus.Debugf("计算结果：延迟时间%f-----%s\n", time.Now().Sub(startTime).Seconds(), result)
	return result
}
