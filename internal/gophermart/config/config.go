package config

import (
	"flag"
	"os"

	"github.com/4aleksei/gmart/internal/common/logger"
)

type Config struct {
	Address string

	DatabaseURI          string
	AccrualSystemAddress string

	Key string

	LCfg logger.Config

	PollInterval int64
	RateLimit    int64
}

const (
	addressDefault              string = ":8090"
	levelDefault                string = "debug"
	databaseURIDefault          string = ""
	accrualSystemAddressDefault string = "localhost:8100"
	keyDefault                  string = "verysecret2"
	pollIntervalDefault         int64  = 2
	rateLimitDefault            int64  = 2
)

func GetConfig() *Config {
	cfg := new(Config)

	flag.StringVar(&cfg.Address, "a", addressDefault, "address and port to run server gopthermart")
	flag.StringVar(&cfg.LCfg.Level, "v", levelDefault, "level of logging")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", accrualSystemAddressDefault, "accrual client`s address and port")

	flag.StringVar(&cfg.DatabaseURI, "d", databaseURIDefault, "database postgres URI")
	flag.Int64Var(&cfg.PollInterval, "i", pollIntervalDefault, "interval bd  request for accrual")
	flag.Int64Var(&cfg.RateLimit, "l", rateLimitDefault, "workers count")
	//	flag.StringVar(&cfg.FilePath, "f", FilePathDefault, "FilePath store")

	//	repository.ReadConfigFlag(&cfg.Repcfg)
	//	readConfigFlag(&cfg.DBcfg)

	flag.StringVar(&cfg.Key, "k", keyDefault, "key for signature")
	flag.Parse()

	//	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
	//		cfg.FilePath = envFilePath
	//	}
	if envKey := os.Getenv("KEY"); cfg.Key == keyDefault && envKey != "" {
		cfg.Key = envKey
	}

	if envRunAddr := os.Getenv("RUN_ADDRESS"); cfg.Address == addressDefault && envRunAddr != "" {
		cfg.Address = envRunAddr
	}

	if envdatabaseURI := os.Getenv("DATABASE_URI"); cfg.DatabaseURI == databaseURIDefault && envdatabaseURI != "" {
		cfg.DatabaseURI = envdatabaseURI
	}

	//	repository.ReadConfigEnv(&cfg.Repcfg)
	//	readConfigEnv(&cfg.DBcfg)

	return cfg
}
