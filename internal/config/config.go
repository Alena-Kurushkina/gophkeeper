// Package config works with configuration variables
// It parse command flags and read environment variables.
// If environment variable is defined, it has highest priority.
// Otherwise flag values are applied.
package config

import (
	"crypto/rand"
	"flag"
	"os"
	"sync"
)

// A Config serves all configuration variables.
type Config struct {
	ServerAddress string
	ConnectionStr string
	DBName string
	SecretKey     []byte
	TokenKey      string
	CertPath      string
	CertKeyPath   string
}

var (
	cfg  *Config
	once sync.Once
)

// InitConfig initialize configuration variables from flags values and environment variables.
func InitConfig() *Config {
	once.Do(
		func() {
			// default values
			cfg = &Config{}

			key := make([]byte, 32) // 256-bit key
			rand.Read(key)
			cfg.SecretKey = key

			tokenKey:=make([]byte, 32) // 256-bit key
			rand.Read(tokenKey)
			cfg.TokenKey=string(tokenKey)

			// get flag values
			flag.StringVar(&cfg.ServerAddress, "a", ":50051", "address of HTTP server")
			flag.StringVar(&cfg.ConnectionStr, "d", "mongodb://localhost:27017", "connection string to database")
			flag.StringVar(&cfg.DBName, "n", "gophkeeper", "database name")
			flag.StringVar(&cfg.CertPath, "c", "/Users/alena/app/tls/practicum_gophkeeper_certs/localhost+2.pem", "path to certificate")
			flag.StringVar(&cfg.CertKeyPath, "k", "/Users/alena/app/tls/practicum_gophkeeper_certs/localhost+2-key.pem", "path to certificate private key")
			flag.Parse()

			// read environment variables
			sa, exists := os.LookupEnv("GOPHKEEPER_SERVER_ADDRESS")
			if exists {
				cfg.ServerAddress = sa
			}

			du, exists := os.LookupEnv("GOPHKEEPER_DATABASE_DSN")
			if exists {
				cfg.ConnectionStr = du
			}

			nu, exists := os.LookupEnv("GOPHKEEPER_DATABASE_NAME")
			if exists {
				cfg.DBName = nu
			}

			cu, exists := os.LookupEnv("GOPHKEEPER_CERT_PATH")
			if exists {
				cfg.CertPath = cu
			}

			ku, exists := os.LookupEnv("GOPHKEEPER_CERT_KEY_PATH")
			if exists {
				cfg.CertKeyPath = ku
			}

			su, exists := os.LookupEnv("GOPHKEEPER_SECRET_KEY")
			if exists {
				cfg.SecretKey = []byte(su)
			}

			tu, exists := os.LookupEnv("GOPHKEEPER_TOKEN_KEY")
			if exists {
				cfg.SecretKey = []byte(tu)
			}
		})
	return cfg
}
