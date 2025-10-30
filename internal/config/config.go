package config

import (
	"flag"
	"os"

	"go.uber.org/zap"
)

type Config struct {
	logger               *zap.SugaredLogger
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	JWTKey               string
}

func ParseFlagsServer(logger *zap.SugaredLogger) Config {
	cfg := Config{}
	cfg.logger = logger

	cfg.parseCommandLineServer()
	cfg.parseEnvironmentServer()

	return cfg
}

func (cfg *Config) parseCommandLineServer() {

	addrTmp := flag.String("a", "localhost:8080", "address and port to run server")
	connStrTmp := flag.String("d", "postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable", "postgress connection string")
	asaTmp := flag.String("r", "", "extrnal system address")
	jwtKey := flag.String("w", "12345", "extrnal system address")

	flag.Parse()

	if addrTmp != nil {
		cfg.RunAddress = *addrTmp
	}
	if connStrTmp != nil {
		cfg.DatabaseURI = *connStrTmp
	}
	if asaTmp != nil {
		cfg.AccrualSystemAddress = *asaTmp
	}
	if jwtKey != nil {
		cfg.JWTKey = *jwtKey
	}

}

func (cfg *Config) parseEnvironmentServer() {
	varAdrHost, ok := os.LookupEnv("RUN_ADDRESS")
	if ok {
		cfg.RunAddress = varAdrHost
	}

	varDatabaseURI, ok := os.LookupEnv("DATABASE_URI")
	if ok {
		cfg.DatabaseURI = varDatabaseURI
	}

	varASA, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if ok {
		cfg.AccrualSystemAddress = varASA
	}

	varJWTK, ok := os.LookupEnv("JWT_KEY")
	if ok {
		cfg.JWTKey = varJWTK
	}
}
