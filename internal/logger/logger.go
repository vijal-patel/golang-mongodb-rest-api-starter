package logger

import (
	"os"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() (*zap.Logger, zap.AtomicLevel) {
	atom := zap.NewAtomicLevel()
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.CallerKey = "caller"
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	), zap.AddCaller())
	defer logger.Sync()
	return logger, atom
}

func GetLoggerForContext(c echo.Context, logger *zap.Logger) *zap.SugaredLogger {
	id := c.Response().Header().Get(echo.HeaderXRequestID)
	childLogger := logger.With(zap.String("requestId", id))
	return childLogger.Sugar()
}
