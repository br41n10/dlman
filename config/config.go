package config

import (
	"log"
	"os"
)

var (
	LocalVendorPath string
	RedisAddr       string
	Sqlite3Path     string

	MirrorCarrier string
	JwtKey        string
)

func init() {
	_ = os.Getenv("DLMAN_ENV")
	LocalVendorPath = GetenvOrExit("DLMAN_LOCAL_VENDOR_PATH")
	RedisAddr = GetenvOrExit("DLMAN_REDIS_ADDR")
	Sqlite3Path = GetenvOrExit("DLMAN_SQLITE3_PATH")

	MirrorCarrier = "LOCAL"
	JwtKey = "guess-guess-guess"
}

func GetenvOrExit(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		log.Fatalf("env variable not define: %s", key)
	}
	return value
}
