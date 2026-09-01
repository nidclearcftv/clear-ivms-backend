package env

import (
	"fmt"
	"os"

	"github.com/Netflix/go-env"
	"github.com/joho/godotenv"
)

func LoadEnv[T any](v T) error {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	if _, err := env.UnmarshalFromEnviron(v); err != nil {
		return err
	}
	return nil
}
