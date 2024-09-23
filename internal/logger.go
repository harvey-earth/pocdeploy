package internal

import (
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the logging instance
var Logger *zap.SugaredLogger

// InitLogger initializes the zap logger
func InitLogger() error {

	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "timestamp"
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder

	config := zap.Config{
		DisableStacktrace: true,
		DisableCaller:     true,
		Encoding:          "console",
		OutputPaths: []string{
			"stdout",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		EncoderConfig: cfg,
	}

	if d := viper.GetBool("debug"); d {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.DisableStacktrace = false
	} else if v := viper.GetBool("verbose"); v {
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else if q := viper.GetBool("quiet"); q {
		config.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	} else {
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}
	defer logger.Sync()
	Logger = logger.Sugar()

	return nil
}

func Debug(message string) {
	Logger.Debug(message)
}

func Info(message string) {
	Logger.Info(message)
}

func Warn(message string) {
	Logger.Warn(message)
}

func Error(err error) {
	Logger.Error(err)
	os.Exit(1)
}

func Fatal(err error) {
	Logger.Fatal(err)
}
