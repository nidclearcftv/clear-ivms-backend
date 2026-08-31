package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type Options struct {
	Level            string `validate:"omitempty,oneof=debug info warn error dpanic panic fatal"`
	Development      bool
	Encoding         string `validate:"omitempty,oneof=json console"`
	OutputPaths      []string
	ErrorOutputPaths []string
	InitialFields    map[string]interface{}
}

// New builds a *zap.SugaredLogger from Options. Level accepts zap's level
// names (debug, info, warn, error, dpanic, panic, fatal) and defaults to
// info when empty. Encoding defaults to "console" when Development is set
// and "json" otherwise.
func New(opts Options) (*zap.SugaredLogger, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	level, err := parseLevel(opts.Level)
	if err != nil {
		return nil, err
	}

	encoding := opts.Encoding
	if encoding == "" {
		if opts.Development {
			encoding = "console"
		} else {
			encoding = "json"
		}
	}

	outputPaths := opts.OutputPaths
	if len(outputPaths) == 0 {
		outputPaths = []string{"stdout"}
	}

	errorOutputPaths := opts.ErrorOutputPaths
	if len(errorOutputPaths) == 0 {
		errorOutputPaths = []string{"stderr"}
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	if opts.Development {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	}
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      opts.Development,
		Encoding:         encoding,
		EncoderConfig:    encoderCfg,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errorOutputPaths,
		InitialFields:    opts.InitialFields,
	}

	log, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("logger: failed to build zap logger: %w", err)
	}

	return log.Sugar(), nil
}

func parseLevel(level string) (zapcore.Level, error) {
	if level == "" {
		return zapcore.InfoLevel, nil
	}

	var l zapcore.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return 0, fmt.Errorf("logger: invalid level %q: %w", level, err)
	}

	return l, nil
}
