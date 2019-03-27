package wallet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	bip32 "github.com/tyler-smith/go-bip32"

	"github.com/sirupsen/logrus"

	"apcchis.com/apcc/common"

	bip44 "github.com/edunuzzi/go-bip44"
	bip39 "github.com/tyler-smith/go-bip39"
	"github.com/tyler-smith/go-bip39/wordlists"
)

const (
	ApccCoinType bip44.CoinType = 334
	ApccCoinTest bip44.CoinType = 335
	//存储钱包集的文件名
	WalletFile = "apcc.wallet"
)

type HDWallet struct {
	Mnemonic        string
	Password        string
	ExternalAddress []Address
	InternalAddress []Address
}
type Address struct {
	PrivKey    []byte
	Purpose    uint32
	CoinType   uint32
	Account    uint32
	ChangeType uint32
	AddrIndex  uint32
	Value      string
}

var MyWallet *HDWallet

//创建钱包
func Create(passwd string) (wallet *HDWallet) {

	if IsWalletExists() {
		logrus.Warnln("创建钱包失败:当前已存在钱包apcc.wallet,如需重新创建请先删除apcc.wallet")
		return
	}

	if len(passwd) != 16 {
		logrus.Fatalln("创建钱包失败:密码必须16位(字母或者数字)")
		return
	}
	common.WalletPassWd = passwd

	bip39.SetWordList(wordlists.ChineseSimplified)
	// 生成随机数
	entropy, _ := bip39.NewEntropy(128)

	// 生成助记词
	mmic, _ := bip39.NewMnemonic(entropy)

	fmt.Println("请牢牢记住以下12个中文助记词,以便于找回账户地址:")
	fmt.Println("#############################################################")
	fmt.Printf("#             %s             #\n", mmic)
	fmt.Println("#############################################################")

	MyWallet = &HDWallet{mmic, common.WalletPassWd, nil, nil}
	MyWallet.ExternalAddress = make([]Address, 0)
	MyWallet.InternalAddress = make([]Address, 0)
	MyWallet.CreateAddress(0, 0, 0)
	log.Printf("您的第一个APCC接收地址为:%s \n", MyWallet.ExternalAddress[0].Value)

	MyWallet.SaveWallets()
	return MyWallet
}

//2.获取钱包地址
func (wallet *HDWallet) CreateAddress(acctIdx, changeType, index uint32) (addr Address) {
	//根据助记词密码生成随机种子
	seed := bip39.NewSeed(wallet.Mnemonic, "")

	masterKey, _ := bip32.NewMasterKey(seed)
	purposeKey, _ := masterKey.NewChildKey(bip32.FirstHardenedChild + common.Purpose)
	coinTypeKey, _ := purposeKey.NewChildKey(bip32.FirstHardenedChild + common.ConinType)
	accountKey, _ := coinTypeKey.NewChildKey(bip32.FirstHardenedChild)

	switch changeType {
	case 0:
		changeKey, _ := accountKey.NewChildKey(changeType)

		addressKey, _ := changeKey.NewChildKey(index)
		pubKey := addressKey.PublicKey().Key
		addr = Address{PrivKey: addressKey.Key, Purpose: common.Purpose, Account: acctIdx, ChangeType: changeType, AddrIndex: index, Value: common.GetAddress(pubKey)}
		wallet.ExternalAddress = append(wallet.ExternalAddress, addr)
	case 1:
		changeKey, _ := accountKey.NewChildKey(changeType)

		addressKey, _ := changeKey.NewChildKey(index)
		addr = Address{PrivKey: addressKey.Key, Purpose: common.Purpose, Account: acctIdx, ChangeType: changeType, AddrIndex: index, Value: common.GetAddress(addressKey.PublicKey().Key)}
		wallet.InternalAddress = append(wallet.InternalAddress, addr)

	}

	//保存
	wallet.SaveWallets()
	return
}

//3.保存钱包集信息到文件
func (wallet *HDWallet) SaveWallets() {

	WalletFile := fmt.Sprintf(WalletFile)

	encoder, err := json.Marshal(wallet)

	if err != nil {

		log.Panic(err)
	}
	xpass, err := common.AesCBCEncrypt(encoder, []byte(wallet.Password))
	if err != nil {
		log.Fatal("保存钱包加密失败\n", err)
		return
	}

	if err := json.Unmarshal(encoder, &wallet); err != nil {
		log.Fatalln("解析JSON失败", err)
	}

	// pass64 := base64.StdEncoding.EncodeToString(xpass)
	// fmt.Printf("加密后:%v\n", pass64)
	// 将序列化以后的数覆盖写入到文件
	err = ioutil.WriteFile(WalletFile, xpass, 0664)
	if err != nil {
		log.Panic(err)
	}
}

//用密码打开钱包
func (wallet *HDWallet) OpenWallet() {
	walletFile, err := ioutil.ReadFile(WalletFile)
	if err != nil {
		logrus.Fatalln("钱包apcc.wallet不存在")
	}

	if len(wallet.Password) != 16 {
		logrus.Fatalln("打开钱包失败:密码必须16位(字母或者数字)")
		return
	}

	tpass, err := common.AesCBCDncrypt(walletFile, []byte(wallet.Password))
	if err != nil {
		log.Fatalln("钱包解密错误")
	}

	if err := json.Unmarshal(tpass, wallet); err != nil {
		log.Fatalln("密码错误或钱包文件被损坏", err)
	}
	MyWallet = wallet
}

//判断数据文件是否存在
func IsWalletExists() bool {

	if _, err := os.Stat(WalletFile); os.IsNotExist(err) {

		return false
	}
	return true
}
