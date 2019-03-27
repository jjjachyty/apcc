package db

import (
	"encoding/hex"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"apcchis.com/apcc/common"
	"github.com/boltdb/bolt"
)

var boltDB *bolt.DB

type KV struct {
	Key   string
	Value []byte
}

func open() {
	if boltDB == nil || boltDB.Path() == "" {
		boltdb, err := bolt.Open(common.DBName, 0600, &bolt.Options{Timeout: 2 * time.Second})
		if err != nil {
			log.Fatal(err)
		}
		boltDB = boltdb
	}
}

func GetTx() *bolt.Tx {
	open()
	if tx, err := boltDB.Begin(true); err == nil {
		return tx
	}
	logrus.Errorln("开启数据库事物失败")
	return nil
}

func GetDB() *bolt.DB {
	open()
	return boltDB
}

func Close() {
	if boltDB != nil && boltDB.Path() != "" {
		boltDB.Close()
	}
}

func Get(bucketName string, key []byte) []byte {

	value := make([]byte, 0)
	open()

	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b != nil {
			value = b.Get(key)
		}
		return nil
	})

	return value
}

func GetBucket(bucketName string, iterator func(key, vals []byte)) {
	open()
	defer boltDB.Close()
	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b != nil {
			c := b.Cursor()
			for k, v := c.First(); len(v) > 0; k, v = c.Next() {
				iterator(k, v)
			}
		}
		return nil
	})

}

func BucketIterator(bucketName string, iterator func(buctet *bolt.Bucket)) {
	open()
	defer boltDB.Close()
	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b != nil {
			iterator(b)
		}
		return nil
	})

}

func DeleteBucket(bucketName string) (err error) {
	open()
	boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))

		// 删除原有UTXO表
		if b != nil {

			if err = tx.DeleteBucket([]byte(bucketName)); err != nil {
				return err
			}

		}
		return nil
	})
	return
}

func CreateBucket(bucketName string) error {
	open()
	boltDB.Update(func(tx *bolt.Tx) error {
		if b, err := tx.CreateBucket([]byte(bucketName)); b != nil {
			return nil
		} else {
			return err
		}

	})
	return nil
}

//Update 更新bucke操作
func Update(bucketName string, key, data []byte) (err error) {
	open()
	return boltDB.Update(func(tx *bolt.Tx) (err error) {
		var b *bolt.Bucket
		if b, err = tx.CreateBucketIfNotExists([]byte(bucketName)); err == nil {
			// key, _ := hex.DecodeString(key)
			if err = b.Put(key, data); err != nil {
				logrus.Fatal(err)
				return
			}
			logrus.Debugf("UPDATE bucket = %s KEY=%x \n\n\n\n", bucketName, key)
		}
		return
	})
}

//Update 更新bucke操作
func UpdateFunc(fc func(tx *bolt.Tx) error) (err error) {
	open()
	defer boltDB.Close()
	return boltDB.Update(func(tx *bolt.Tx) (err error) {
		return fc(tx)
	})
}

//Update 更新bucke操作
func GetFunc(fc func(tx *bolt.Tx) error) (err error) {
	open()
	defer boltDB.Close()
	return boltDB.View(func(tx *bolt.Tx) (err error) {
		return fc(tx)
	})
}

//UpdateBatch 批量更新bucke操作
func UpdateBatch(bucketName string, data []KV) (err error) {
	open()

	return boltDB.Batch(func(tx *bolt.Tx) (err error) {
		var b *bolt.Bucket
		if b, err = tx.CreateBucketIfNotExists([]byte(bucketName)); err == nil {
			for _, obj := range data {
				keyByts, _ := hex.DecodeString(obj.Key)
				if err = b.Put(keyByts, obj.Value); err != nil {
					log.Fatalf("批量更新失败%s\n", err)
					return
				}

			}

		}
		return
	})

}

//判断数据文件是否存在
func IsDBExists(dbName string) bool {

	if _, err := os.Stat(dbName); os.IsNotExist(err) {

		return false
	}
	return true
}
