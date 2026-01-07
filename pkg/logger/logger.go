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
	"sync"
)

var (
	globalLogger *DingoLogger
	once         sync.Once
)

func newDingoLogger(opts ...Option) *DingoLogger {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	zapLogger := newZapLogger(cfg)
	sugar := zapLogger.Sugar()

	return &DingoLogger{
		zapLogger: zapLogger,
		sugar:     sugar,
	}
}

func InitGlobalLogger(opts ...Option) *DingoLogger {
	once.Do(func() {
		globalLogger = newDingoLogger(opts...)
	})
	return globalLogger
}

func GetLogger() *DingoLogger {
	once.Do(func() {
		if globalLogger == nil {
			globalLogger = newDingoLogger()
		}
	})
	return globalLogger
}

func Debug(message string) {
	GetLogger().Debug(message)
}

func Info(message string) {
	GetLogger().Info(message)
}

func Warn(message string) {
	GetLogger().Warn(message)
}

func Error(message string) {
	GetLogger().Error(message)
}

func Fatal(message string) {
	GetLogger().Fatal(message)
}

func Panic(message string) {
	GetLogger().Panic(message)
}

func Debugf(message string, args ...interface{}) {
	GetLogger().Debugf(message, args...)
}

func Infof(message string, args ...interface{}) {
	GetLogger().Infof(message, args...)
}

func Warnf(message string, args ...interface{}) {
	GetLogger().Warnf(message, args...)
}

func Errorf(message string, args ...interface{}) {
	GetLogger().Errorf(message, args...)
}

func Fatalf(message string, args ...interface{}) {
	GetLogger().Fatalf(message, args...)
}

func Panicf(message string, args ...interface{}) {
	GetLogger().Panicf(message, args...)
}

func Sync() error {
	return GetLogger().Sync()
}
