package application

import (
	"fmt"
	"os"
	"strings"
	"time"

	logcore "github.com/surge-go/fox/core/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const loggerTimeEncodingDatetime = "datetime"

// initLogger 初始化应用日志实例。
func (app *Application) initLogger() error {
	if app.cfg == nil || app.cfg.Logger == nil {
		return nil
	}

	logger, err := app.newLogger()
	if err != nil {
		return err
	}
	app.logger = logger
	app.restoreLogger = zap.ReplaceGlobals(logger)
	return nil
}

func (app *Application) newLogger() (*zap.Logger, error) {
	if app.cfg.Logger.Encoder != nil && app.cfg.Logger.Encoder.TimeEncoding == loggerTimeEncodingDatetime {
		return newDatetimeLogger(app.cfg.Logger)
	}

	return logcore.New(app.toLoggerConfig())
}

func newDatetimeLogger(cfg *LoggerConfig) (*zap.Logger, error) {
	level, err := parseLoggerLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	encoder, err := buildDatetimeLoggerEncoder(cfg.Format, cfg.Encoder)
	if err != nil {
		return nil, err
	}

	writer, err := buildLoggerWriter(cfg)
	if err != nil {
		return nil, err
	}

	core := zapcore.NewCore(encoder, writer, level)
	if cfg.Sampling != nil && cfg.Sampling.Enabled {
		core = zapcore.NewSamplerWithOptions(core, time.Second, cfg.Sampling.Initial, cfg.Sampling.Thereafter)
	}

	options, err := buildLoggerOptions(cfg)
	if err != nil {
		return nil, err
	}
	return zap.New(core, options...), nil
}

func parseLoggerLevel(level string) (zapcore.Level, error) {
	if level == "" {
		return zapcore.InfoLevel, nil
	}

	var parsed zapcore.Level
	if err := parsed.Set(level); err != nil {
		return zapcore.InfoLevel, fmt.Errorf("unsupported logger level %q", level)
	}
	return parsed, nil
}

func buildDatetimeLoggerEncoder(format string, cfg *LoggerEncoderConfig) (zapcore.Encoder, error) {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "ts",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     datetimeTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg != nil {
		applyLoggerEncoderConfig(&encoderConfig, cfg)
		if err := applyLoggerEncoderParsers(&encoderConfig, cfg); err != nil {
			return nil, err
		}
	}

	switch format {
	case "", "json":
		return zapcore.NewJSONEncoder(encoderConfig), nil
	case "console":
		return zapcore.NewConsoleEncoder(encoderConfig), nil
	default:
		return nil, fmt.Errorf("unsupported logger format %q", format)
	}
}

func applyLoggerEncoderConfig(encoderConfig *zapcore.EncoderConfig, cfg *LoggerEncoderConfig) {
	if cfg.MessageKey != "" {
		encoderConfig.MessageKey = cfg.MessageKey
	}
	if cfg.LevelKey != "" {
		encoderConfig.LevelKey = cfg.LevelKey
	}
	if cfg.TimeKey != "" {
		encoderConfig.TimeKey = cfg.TimeKey
	}
	if cfg.NameKey != "" {
		encoderConfig.NameKey = cfg.NameKey
	}
	if cfg.CallerKey != "" {
		encoderConfig.CallerKey = cfg.CallerKey
	}
	encoderConfig.FunctionKey = cfg.FunctionKey
	if cfg.StacktraceKey != "" {
		encoderConfig.StacktraceKey = cfg.StacktraceKey
	}
	if cfg.LineEnding != "" {
		encoderConfig.LineEnding = cfg.LineEnding
	}
}

func applyLoggerEncoderParsers(encoderConfig *zapcore.EncoderConfig, cfg *LoggerEncoderConfig) error {
	if cfg.DurationEncoding != "" {
		encoder, err := parseLoggerDurationEncoder(cfg.DurationEncoding)
		if err != nil {
			return err
		}
		encoderConfig.EncodeDuration = encoder
	}
	if cfg.LevelEncoding != "" {
		encoder, err := parseLoggerLevelEncoder(cfg.LevelEncoding)
		if err != nil {
			return err
		}
		encoderConfig.EncodeLevel = encoder
	}
	return nil
}

func datetimeTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func parseLoggerDurationEncoder(encoding string) (zapcore.DurationEncoder, error) {
	switch encoding {
	case "seconds":
		return zapcore.SecondsDurationEncoder, nil
	case "millis":
		return zapcore.MillisDurationEncoder, nil
	case "nanos":
		return zapcore.NanosDurationEncoder, nil
	case "string":
		return zapcore.StringDurationEncoder, nil
	default:
		return nil, fmt.Errorf("unsupported logger duration_encoding %q", encoding)
	}
}

func parseLoggerLevelEncoder(encoding string) (zapcore.LevelEncoder, error) {
	switch encoding {
	case "lowercase":
		return zapcore.LowercaseLevelEncoder, nil
	case "capital":
		return zapcore.CapitalLevelEncoder, nil
	case "color":
		return zapcore.CapitalColorLevelEncoder, nil
	default:
		return nil, fmt.Errorf("unsupported logger level_encoding %q", encoding)
	}
}

func buildLoggerWriter(cfg *LoggerConfig) (zapcore.WriteSyncer, error) {
	switch cfg.Output {
	case "", "stdout":
		return zapcore.Lock(os.Stdout), nil
	case "stderr":
		return zapcore.Lock(os.Stderr), nil
	case "file":
		if strings.TrimSpace(cfg.File) == "" {
			return nil, fmt.Errorf("logger file must not be empty when output is file")
		}
		return zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    loggerRotationMaxSize(cfg.Rotation),
			MaxAge:     loggerRotationMaxAge(cfg.Rotation),
			MaxBackups: loggerRotationMaxBackups(cfg.Rotation),
			LocalTime:  loggerRotationLocalTime(cfg.Rotation),
			Compress:   loggerRotationCompress(cfg.Rotation),
		}), nil
	default:
		return nil, fmt.Errorf("unsupported logger output %q", cfg.Output)
	}
}

func buildLoggerOptions(cfg *LoggerConfig) ([]zap.Option, error) {
	options := []zap.Option{zap.ErrorOutput(buildLoggerErrorOutput(cfg.ErrorOutput))}
	if cfg.Development {
		options = append(options, zap.Development())
	}
	if cfg.AddCaller {
		options = append(options, zap.AddCaller())
	}
	if cfg.CallerSkip > 0 {
		options = append(options, zap.AddCallerSkip(cfg.CallerSkip))
	}
	if cfg.StacktraceLevel != "" && cfg.StacktraceLevel != "none" {
		level, err := parseLoggerLevel(cfg.StacktraceLevel)
		if err != nil {
			return nil, fmt.Errorf("unsupported logger stacktrace_level %q", cfg.StacktraceLevel)
		}
		options = append(options, zap.AddStacktrace(level))
	}
	if len(cfg.InitialFields) > 0 {
		fields := make([]zap.Field, 0, len(cfg.InitialFields))
		for key, value := range cfg.InitialFields {
			fields = append(fields, zap.String(key, value))
		}
		options = append(options, zap.Fields(fields...))
	}
	return options, nil
}

func buildLoggerErrorOutput(output string) zapcore.WriteSyncer {
	switch strings.TrimSpace(output) {
	case "", "stderr":
		return zapcore.Lock(os.Stderr)
	case "stdout":
		return zapcore.Lock(os.Stdout)
	default:
		return zapcore.AddSync(&lumberjack.Logger{Filename: output})
	}
}

func loggerRotationMaxSize(cfg *LoggerRotationConfig) int {
	if cfg == nil {
		return 0
	}
	return cfg.MaxSize
}

func loggerRotationMaxAge(cfg *LoggerRotationConfig) int {
	if cfg == nil {
		return 0
	}
	return cfg.MaxAge
}

func loggerRotationMaxBackups(cfg *LoggerRotationConfig) int {
	if cfg == nil {
		return 0
	}
	return cfg.MaxBackups
}

func loggerRotationLocalTime(cfg *LoggerRotationConfig) bool {
	return cfg != nil && cfg.LocalTime
}

func loggerRotationCompress(cfg *LoggerRotationConfig) bool {
	return cfg != nil && cfg.Compress
}

// toLoggerConfig 将应用配置转换为 fox/core/logger 配置。
func (app *Application) toLoggerConfig() *logcore.Config {
	if app.cfg == nil || app.cfg.Logger == nil {
		return nil
	}

	cfg := app.cfg.Logger
	return &logcore.Config{
		Level:           logcore.Level(cfg.Level),
		Format:          logcore.Format(cfg.Format),
		Output:          logcore.Output(cfg.Output),
		File:            cfg.File,
		ErrorOutput:     cfg.ErrorOutput,
		Development:     cfg.Development,
		AddCaller:       cfg.AddCaller,
		CallerSkip:      cfg.CallerSkip,
		StacktraceLevel: logcore.StacktraceLevel(cfg.StacktraceLevel),
		InitialFields:   cfg.InitialFields,
		Encoder:         toLoggerEncoderConfig(cfg.Encoder),
		Rotation:        toLoggerRotationConfig(cfg.Rotation),
		Sampling:        toLoggerSamplingConfig(cfg.Sampling),
	}
}

func toLoggerEncoderConfig(cfg *LoggerEncoderConfig) *logcore.EncoderConfig {
	if cfg == nil {
		return nil
	}

	return &logcore.EncoderConfig{
		MessageKey:       cfg.MessageKey,
		LevelKey:         cfg.LevelKey,
		TimeKey:          cfg.TimeKey,
		NameKey:          cfg.NameKey,
		CallerKey:        cfg.CallerKey,
		FunctionKey:      cfg.FunctionKey,
		StacktraceKey:    cfg.StacktraceKey,
		LineEnding:       cfg.LineEnding,
		TimeEncoding:     cfg.TimeEncoding,
		DurationEncoding: cfg.DurationEncoding,
		LevelEncoding:    cfg.LevelEncoding,
	}
}

func toLoggerRotationConfig(cfg *LoggerRotationConfig) *logcore.RotationConfig {
	if cfg == nil {
		return nil
	}

	return &logcore.RotationConfig{
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
		LocalTime:  cfg.LocalTime,
		Compress:   cfg.Compress,
	}
}

func toLoggerSamplingConfig(cfg *LoggerSamplingConfig) *logcore.SamplingConfig {
	if cfg == nil {
		return nil
	}

	return &logcore.SamplingConfig{
		Enabled:    cfg.Enabled,
		Initial:    cfg.Initial,
		Thereafter: cfg.Thereafter,
	}
}
