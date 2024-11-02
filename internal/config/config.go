package config

type Config struct {
	MemStoreAddr   string // address of the RESP memory backend, like Redis
	MaxConnections int
}

func NewConfigFromDotenv(fp string) (Config, error) {
	cfg, err := ReadDotenv(fp)
	if err != nil {
		return Config{}, err
	}

	return Config{MemStoreAddr: cfg["SECURE_STORE_ADDRESS"]}, nil
}
