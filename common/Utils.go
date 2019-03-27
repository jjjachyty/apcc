package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

	"github.com/btcsuite/golangcrypto/ripemd160"
)

//将int64转换为bytes
func IntToHex(num int64) []byte {

	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {

		log.Panic(err)
	}

	return buff.Bytes()
}

// 标准的JSON字符串转数组
func Json2Array(jsonString string) []string {

	//json 到 []string
	var sArr []string
	if err := json.Unmarshal([]byte(jsonString), &sArr); err != nil {

		log.Panic(err)
	}
	return sArr
}

// 字节数组反转
func ReverseBytes(data []byte) {

	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {

		data[i], data[j] = data[j], data[i]
	}
}

// 将结构体序列化成字节数组
func GobEncode(data interface{}) []byte {

	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func CommandToBytes(command string) []byte {

	// 消息在底层就是字节序列,前12个字节指定了命令名（比如这里的 version）
	var bytes [COMMAND_LENGTH]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func BytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

//将公钥进行两次哈希
func Ripemd160Hash(publicKey []byte) []byte {

	//1.hash256
	hash256 := sha256.New()
	hash256.Write(publicKey)
	hash := hash256.Sum(nil)

	//2.ripemd160
	ripemd160 := ripemd160.New()
	ripemd160.Write(hash)

	return ripemd160.Sum(nil)
}

//两次sha256哈希生成校验和
func CheckSum(bytes []byte) []byte {

	//hasher := sha256.New()
	//hasher.Write(bytes)
	//hash := hasher.Sum(nil)
	//与下面一句等同
	//hash := sha256.Sum256(bytes)

	hash1 := sha256.Sum256(bytes)
	hash2 := sha256.Sum256(hash1[:])

	return hash2[:AddressChecksumLen]
}

func GetAddress(pubKeyByts []byte) string {

	//1.使用RIPEMD160(SHA256(PubKey)) 哈希算法，取公钥并对其哈希两次
	ripemd160Hash := Ripemd160Hash(pubKeyByts)

	//2.拼接版本
	version_ripemd160Hash := append([]byte{0x00}, ripemd160Hash...)
	//3.两次sha256生成校验和
	checkSumBytes := CheckSum(version_ripemd160Hash)
	//4.拼接校验和
	bytes := append(version_ripemd160Hash, checkSumBytes...)

	return string(Base58Encode(bytes))
}
