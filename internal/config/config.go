package config

import (
	"fmt"
	"strconv"
)

const DEFAULT_MAX_CONN int = 4

type Config struct {
	MemStoreAddr   string // address of the RESP memory backend, like Redis
	MaxConnections int
}

func NewConfigFromDotenv(fp string) (Config, error) {
	cfg, err := ReadDotenv(fp)
	if err != nil {
		return Config{}, err
	}

	maxConn, err := strconv.Atoi(cfg["MAX_SERVER_CONNECTIONS"])
	if err != nil {
		fmt.Println("[warning] issue getting MAX_SERVER_CONNECTIONS, using default value", DEFAULT_MAX_CONN)
		maxConn = DEFAULT_MAX_CONN
	}

	return Config{MemStoreAddr: cfg["SECURE_STORE_ADDRESS"],
		MaxConnections: maxConn}, nil
}
