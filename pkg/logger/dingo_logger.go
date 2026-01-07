// Copyright (c) 2025 dingodb.com, Inc. All Rights Reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type DingoLogger struct {
	zapLogger *zap.Logger
	sugar     *zap.SugaredLogger
}

func convertToLevel(loglevel string) zapcore.Level {
	var level zapcore.Level
	switch loglevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "fatal":
		level = zap.FatalLevel
	case "panic":
		level = zap.PanicLevel
	default:
		level = zap.InfoLevel
	}

	return level
}

func newZapLogger(cfg *logConfig) *zap.Logger {
	hook := lumberjack.Logger{
		Filename:   cfg.LogFile,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	writeSyncer := zapcore.AddSync(&hook)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	level := convertToLevel(cfg.LogLevel)
	core := zapcore.NewCore(
		encoder,
		writeSyncer,
		level,
	)

	return zap.New(core)
}

func (logger *DingoLogger) Info(message string) {
	logger.zapLogger.Info(message)
}

func (logger *DingoLogger) Debug(message string) {
	logger.zapLogger.Debug(message)
}

func (logger *DingoLogger) Error(message string) {
	logger.zapLogger.Error(message)
}

func (logger *DingoLogger) Warn(message string) {
	logger.zapLogger.Warn(message)
}

func (logger *DingoLogger) Fatal(message string) {
	logger.zapLogger.Fatal(message)
}

func (logger *DingoLogger) Panic(message string) {
	logger.zapLogger.Panic(message)
}

func (logger *DingoLogger) Infof(message string, args ...interface{}) {
	logger.sugar.Infof(message, args...)
}

func (logger *DingoLogger) Debugf(message string, args ...interface{}) {
	logger.sugar.Debugf(message, args...)
}

func (logger *DingoLogger) Warnf(template string, args ...interface{}) {
	logger.zapLogger.Sugar().Warnf(template, args...)
}

func (logger *DingoLogger) Errorf(template string, args ...interface{}) {
	logger.zapLogger.Sugar().Errorf(template, args...)
}

func (logger *DingoLogger) Fatalf(template string, args ...interface{}) {
	logger.zapLogger.Sugar().Fatalf(template, args...)
}

func (logger *DingoLogger) Panicf(template string, args ...interface{}) {
	logger.zapLogger.Sugar().Panicf(template, args...)
}

func (logger *DingoLogger) Sync() error {
	return logger.zapLogger.Sync()
}
