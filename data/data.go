package data

import (
	"database/sql"
	"dlman/config"
	"github.com/go-redis/redis/v7"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var Db *sql.DB
var RedisCli *redis.Client

type NullString sql.NullString

func init() {
	var err error

	// 设置sqlite3数据库连接
	Db, err = sql.Open("sqlite3", config.Sqlite3Path)
	if err != nil {
		log.Fatal(err)
	}

	RedisCli = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := RedisCli.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(pong)
}

// TODO: 有空了模仿 sql.NullString 自造一下 uuid 的
// Scan implements the Scanner interface.
//func (su *sqlUUID) Scan(value interface{}) error {
//
//}
