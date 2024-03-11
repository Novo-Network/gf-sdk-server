package loger

import (
	"io"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConf struct {
	Level   string `yaml:"level"`
	Path    string `yaml:"path"`
	Size    int    `yaml:"size"`
	Backups int    `yaml:"backups"`
}

var levelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel, //
	"error": zapcore.ErrorLevel,
	"panic": zapcore.PanicLevel,
	"fatal": zapcore.FatalLevel,
}

var Logger *zap.SugaredLogger

func InitLogger() {
	Logger = InitZapLogger("loger/app.loger", "info", 500, 20)
}

func InitZapLogger(fileName string, level string, size int, backups int) *zap.SugaredLogger {
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		TimeKey:     "ts",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		CallerKey:    "file",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})

	rotateLogger := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    size,
		MaxBackups: backups,
		LocalTime:  true,
	}

	logLevel := levelMap[strings.ToLower(level)]
	ioWriter := io.MultiWriter(rotateLogger, os.Stdout)
	core := zapcore.NewCore(encoder, zapcore.AddSync(ioWriter), logLevel)
	l := zap.New(core, zap.AddCaller())
	return l.Sugar()
}
