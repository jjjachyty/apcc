package core

import (
	"crypto/elliptic"
	"testing"

	"github.com/btcsuite/btcd/btcec"

	// "github.com/btcsuite/btcd/txscript"

	// "github.com/btcsuite/btcutil"

	// "golang.org/x/crypto/ripemd160"

	// "apcchis.com/apcc/common"

	"apcchis.com/apcc/common"
	bip32 "github.com/FactomProject/go-bip32"
	bip39 "github.com/tyler-smith/go-bip39"
	hdwallet "github.com/wemeetagain/go-hdwallet"
	// "github.com/tyler-smith/go-bip39/wordlists"
)

// func TestTXInput_UnlockWithAddress(t *testing.T) {
// 	// hash160 := []byte("04ab90bdcd11c503020554fc325049c71fbd4d1d")
// 	// fmt.Printf("%s\n", base58.CheckEncode(hash160[:ripemd160.Size], byte(334)))
// 	btcutil.NewTx()
// 	bip39.SetWordList(wordlists.ChineseSimplified)
// 	// 生成随机数
// 	entropy, _ := bip39.NewEntropy(128)

// 	// 生成助记词
// 	mmic, _ := bip39.NewMnemonic(entropy)

// 	fmt.Println("请牢牢记住以下12个中文助记词,以便于找回账户地址:")
// 	fmt.Println("#############################################################")
// 	fmt.Printf("#             %s             #\n", mmic)
// 	fmt.Println("#############################################################")
// 	seed := bip39.NewSeed(mmic, "")
// 	xKey, _ := bip44.NewKeyFromSeedBytes(seed, bip44.MAINNET)

// 	accountKey, _ := xKey.BIP44AccountKey(334, 0, true)

// 	addr, _ := accountKey.DeriveP2PKAddress(bip44.ExternalChangeType, 0, bip44.MAINNET)

// 	script, _ := txscript.PayToAddrScript(addr)

// 	fmt.Printf("txscript.PayToAddrScript(addr)=")
// 	fmt.Printf("addr.pk = %x\n", addr.PrivKey)
// 	pk, _ := btcec.PrivKeyFromBytes(elliptic.P256(), addr.PrivKey)
// 	pkByts := bytes.Join([][]byte{pk.X.Bytes(), pk.Y.Bytes()}, []byte{})
// 	puk := bytes.Join([][]byte{pk.PublicKey.X.Bytes(), pk.PublicKey.Y.Bytes()}, []byte{})
// 	fmt.Printf("PUK end = %x\n", puk)
// 	//1.hash256
// 	hash256 := sha256.New()
// 	hash256.Write(pkByts)
// 	hash := hash256.Sum(nil)
// 	fmt.Printf("HASH PUK %x\n", hash)
// 	//2.ripemd160
// 	ripemd160 := ripemd160.New()
// 	ripemd160.Write(pkByts)
// 	fmt.Printf("ripemd160 PUK %x\n", ripemd160.Sum(nil))
// 	fmt.Printf("btcutil.Hash160=%x", btcutil.Hash160(puk))
// 	fmt.Printf("publickey=%x-\n-address=%x\n", puk, common.Ripemd160Hash(puk))

// 	fmt.Printf("Base58 address %x\n", common.Base58Decode([]byte(addr.Value)))
// }
func TestHD(t *testing.T) {
	mnemonic := "yellow yellow yellow yellow yellow yellow yellow yellow yellow yellow yellow yellow"

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		panic(err)
	}
	// Create a master private key
	masterprv := hdwallet.MasterKey(seed)
	t.Logf("masterprv=%s", masterprv.String())
	// Convert a private key to public key
	masterpub := masterprv.Pub()

	// Generate new child key based on private or public key
	// childprv, err := masterprv.Child(0)
	childpub, err := masterpub.Child(0)

	// Create bitcoin address from public key
	address := childpub.Address()
	t.Logf("address=%s", address)
	// Convenience string -> string Child and ToAddress functions
	walletstring := childpub.String()
	childstring, err := hdwallet.StringChild(walletstring, 0)
	t.Logf("%s", childstring)
	childaddress, err := hdwallet.StringAddress(childstring)
	t.Logf("%s", childaddress)
}
func TestNewKeyFromMasterKey(t *testing.T) {
	mnemonic := "yellow yellow yellow yellow yellow yellow yellow yellow yellow yellow yellow yellow"

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		panic(err)
	}

	masterKey, err := bip32.NewMasterKey(seed)
	t.Logf("masterKey=%v", masterKey.String())
	if err != nil {
		panic(err)
	}

	child, err := masterKey.NewChildKey(bip32.FirstHardenedChild + 44)
	if err != nil {
		panic(err)
	}
	t.Logf("44=%v", child.String())
	child, err = child.NewChildKey(bip32.FirstHardenedChild + 0)
	if err != nil {
		panic(err)
	}
	t.Logf("coin=%v", child.String())
	child, err = child.NewChildKey(bip32.FirstHardenedChild)
	if err != nil {
		panic(err)
	}
	t.Logf("account=%v %x", child.String(), child.Depth)
	child, err = child.NewChildKey(0)
	if err != nil {
		panic(err)
	}
	t.Logf("change=%v", child.String())
	t.Logf("change publikey=%v", child.PublicKey().String())

	child, err = child.NewChildKey(1)
	if err != nil {
		panic(err)
	}

	t.Logf("address_index pk =%v", child.String())

	t.Logf("address_index pubkey%x", len(child.PublicKey().Key))

	//1.使用RIPEMD160(SHA256(PubKey)) 哈希算法，取公钥并对其哈希两次
	ripemd160Hash := common.Ripemd160Hash(child.PublicKey().Key)
	t.Logf("address1 %s", common.Base58Encode(ripemd160Hash))

	//2.拼接版本
	version_ripemd160Hash := append([]byte{0x00}, ripemd160Hash...)
	//3.两次sha256生成校验和
	checkSumBytes := common.CheckSum(version_ripemd160Hash)
	//4.拼接校验和
	bytes := append(version_ripemd160Hash, checkSumBytes...)

	t.Logf("address %s", common.Base58Encode(bytes))

}

func TestRsa(t *testing.T) {
	pk, _ := btcec.PrivKeyFromBytes(elliptic.P256(), []byte("xprvA1MQ7TKj2GbtD7QkuaXc22EjQXbB3BgzSHsib9diYjBf6SaZyPd5uhuK2WdhmdNxaThgTh8bbLoZ8BXfGJS8vkkiaRFgPKt6ciMjSn2gBZw"))
	ecdsaPK := pk.ToECDSA()
	t.Logf("address_index pkey=%x %x", ecdsaPK.X.Bytes(), ecdsaPK.Y.Bytes())

	t.Logf("address_index pubkey=%x %x", ecdsaPK.PublicKey.X.Bytes(), ecdsaPK.PublicKey.Y.Bytes())

}
