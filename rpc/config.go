package rpc

import "time"

type Config struct {
	Port         string
	DataDir      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func DefaultConfig() Config {
	return Config{
		Port:         "8080",
		DataDir:      ".",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}
