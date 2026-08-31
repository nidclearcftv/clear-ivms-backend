package main

import (
	"fmt"
	"os"

	"github.com/Netflix/go-env"
	"github.com/joho/godotenv"
)

// Env holds the standalone command's runtime configuration, sourced from
// environment variables (optionally loaded from a .env file first).
type Env struct {
	CMSV6URL      string `env:"CMSV6_URL,required=true"`
	CMSV6Username string `env:"CMSV6_USERNAME,required=true"`
	CMSV6Password string `env:"CMSV6_PASSWORD,required=true"`
}

func loadEnv() (Env, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Env{}, fmt.Errorf("failed to load .env file: %w", err)
	}

	var e Env
	if _, err := env.UnmarshalFromEnviron(&e); err != nil {
		return Env{}, err
	}
	return e, nil
}
