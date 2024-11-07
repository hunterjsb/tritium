package config

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	DEFAULT_MAX_CONN int = 4
	DEFAULT_REPLICAS     = "" // Empty string means no replicas
)

type Config struct {
	MemStoreAddr   string // address of the RESP memory backend, like Redis
	MaxConnections int
	ReplicaAddrs   []string // Optional replica addresses
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

	// Parse replica addresses from env var (comma-separated)
	var replicaAddrs []string
	if replicas := cfg["SECURE_STORE_REPLICAS"]; replicas != "" {
		replicaAddrs = strings.Split(replicas, ",")
		// Trim any whitespace
		for i := range replicaAddrs {
			replicaAddrs[i] = strings.TrimSpace(replicaAddrs[i])
		}
	}

	return Config{
		MemStoreAddr:   cfg["SECURE_STORE_ADDRESS"],
		MaxConnections: maxConn,
		ReplicaAddrs:   replicaAddrs,
	}, nil
}
