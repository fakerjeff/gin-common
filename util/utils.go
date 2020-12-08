package util

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"golang.org/x/crypto/bcrypt"
)

func GetDefaultEnv(key string, defaultValue string) string {
	result := os.Getenv(key)
	if result == "" {
		return defaultValue
	}
	return result
}

// 加密密码
func HashAndSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// 验证密码
func CompareHash(hashedStr string, plainStr []byte) bool {
	byteHash := []byte(hashedStr)

	err := bcrypt.CompareHashAndPassword(byteHash, plainStr)
	if err != nil {
		return false
	}
	return true
}

func PanicError(err error) {
	if err != nil {
		panic(err)
	}
}

func GetLogger(logLevel zapcore.Level, out io.Writer, options ...zap.Option) *zap.Logger {

	syncer := zapcore.AddSync(out)
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), syncer, logLevel)
	return zap.New(core, options...)
}

var logLevelMap = map[string]zapcore.Level{
	"info":  zap.InfoLevel,
	"debug": zap.DebugLevel,
	"warn":  zap.WarnLevel,
	"error": zap.ErrorLevel,
}

func GetLogLevel(level string) zapcore.Level {
	if logLevel, ok := logLevelMap[level]; ok {
		return logLevel
	} else {
		return zap.ErrorLevel
	}
}

func IsBool(val interface{}) (bool, error) {
	switch val := val.(type) {
	case int64:
		return val != 0, nil
	case string:
		return strconv.ParseBool(val)
	default:
		err := fmt.Errorf("unexpected type=%T for Bool", val)
		return false, err
	}
}
