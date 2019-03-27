package db

type DB interface {
	Get(dbName, key string) []byte
	Update(dbName, key string, data []byte) error
}
