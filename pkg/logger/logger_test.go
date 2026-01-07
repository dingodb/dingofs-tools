// Copyright (c) 2025 dingodb.com, Inc. All Rights Reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package logger

import (
	"bytes"
	"sync"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestInitGlobalLogger(t *testing.T) {
	t.Run("TestLoggerInit", func(t *testing.T) {
		globalLogger = nil
		once = sync.Once{}

		logger := InitGlobalLogger(WithLogLevel("debug"))
		assert.NotNil(t, logger)
		assert.Equal(t, logger, globalLogger)

		logger2 := InitGlobalLogger(WithLogLevel("error"))
		assert.Equal(t, logger, logger2)
	})

	t.Run("TestLoggerConcurrentInit", func(t *testing.T) {
		globalLogger = nil
		once = sync.Once{}

		var wg sync.WaitGroup
		loggers := make([]*DingoLogger, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				loggers[index] = InitGlobalLogger()
			}(i)
		}
		wg.Wait()

		// all goroutine have the same logger
		firstLogger := loggers[0]
		for i := 1; i < 10; i++ {
			assert.Equal(t, firstLogger, loggers[i])
		}
	})
}

func TestGetLogger(t *testing.T) {
	t.Run("TestGetLogger", func(t *testing.T) {
		globalLogger = nil
		once = sync.Once{}

		logger := GetLogger()
		assert.NotNil(t, logger)
		assert.Equal(t, logger, globalLogger)
	})

	t.Run("GetLoggerAfterInit", func(t *testing.T) {
		globalLogger = nil
		once = sync.Once{}

		initLogger := InitGlobalLogger(WithLogLevel("warn"))
		getLogger := GetLogger()

		assert.Equal(t, initLogger, getLogger)
	})
}

func TestLogMethods(t *testing.T) {
	core, recorded := observer.New(zap.DebugLevel)
	testLogger := &DingoLogger{
		zapLogger: zap.New(core),
		sugar:     zap.New(core).Sugar(),
	}

	originalLogger := globalLogger
	globalLogger = testLogger
	defer func() { globalLogger = originalLogger }()

	t.Run("TestDebugMethod", func(t *testing.T) {
		Debug("test debug message")
		assert.Equal(t, 1, recorded.Len())
		entry := recorded.All()[0]
		assert.Equal(t, "test debug message", entry.Message)
		assert.Equal(t, zapcore.DebugLevel, entry.Level)
		recorded.TakeAll()
	})

	t.Run("TestInfoMethod", func(t *testing.T) {
		Info("test info message")
		assert.Equal(t, 1, recorded.Len())
		entry := recorded.All()[0]
		assert.Equal(t, "test info message", entry.Message)
		assert.Equal(t, zapcore.InfoLevel, entry.Level)
		recorded.TakeAll()
	})

	t.Run("TestInfofMethod", func(t *testing.T) {
		Infof("user %s logged in with id %d", "dingo", 123)
		assert.Equal(t, 1, recorded.Len())
		entry := recorded.All()[0]
		assert.Equal(t, "user dingo logged in with id 123", entry.Message)
		assert.Equal(t, zapcore.InfoLevel, entry.Level)
		recorded.TakeAll()
	})

	t.Run("TestErrorMethod", func(t *testing.T) {
		Error("test error message")
		assert.Equal(t, 1, recorded.Len())
		entry := recorded.All()[0]
		assert.Equal(t, "test error message", entry.Message)
		assert.Equal(t, zapcore.ErrorLevel, entry.Level)
		recorded.TakeAll()
	})
}

func TestSync(t *testing.T) {
	t.Run("TestSyncMethod", func(t *testing.T) {
		var buf bytes.Buffer
		encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zap.DebugLevel)
		testLogger := &DingoLogger{
			zapLogger: zap.New(core),
			sugar:     zap.New(core).Sugar(),
		}

		originalLogger := globalLogger
		globalLogger = testLogger
		defer func() { globalLogger = originalLogger }()

		err := Sync()
		assert.NoError(t, err)
	})
}

func TestDingoLoggerMethods(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zap.DebugLevel)
	logger := &DingoLogger{
		zapLogger: zap.New(core),
		sugar:     zap.New(core).Sugar(),
	}

	t.Run("TestDingoLoggerDebug", func(t *testing.T) {
		buf.Reset()
		logger.Debug("debug test")
		assert.Contains(t, buf.String(), "debug test")
	})

	t.Run("TestDingoLoggerInfof", func(t *testing.T) {
		buf.Reset()
		logger.Infof("formatted %s", "test")
		assert.Contains(t, buf.String(), "formatted test")
	})

	t.Run("TestDingoLoggerError", func(t *testing.T) {
		buf.Reset()
		logger.Error("error test")
		assert.Contains(t, buf.String(), "error test")
	})
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, DEFAULT_LOG_FILE, cfg.LogFile)
	assert.Equal(t, DEFAULT_LOG_LEVEL, cfg.LogLevel)
	assert.Equal(t, DEFAULT_LOG_FORMAT, cfg.Format)
	assert.False(t, cfg.Stdout)
}

func TestLogOption(t *testing.T) {
	t.Run("TestLoglevelOption", func(t *testing.T) {
		cfg := defaultConfig()
		opt := WithLogLevel("debug")
		opt(cfg)
		assert.Equal(t, "debug", cfg.LogLevel)
	})

	t.Run("TestLogfmtOption", func(t *testing.T) {
		cfg := defaultConfig()
		opt := WithFormat("json")
		opt(cfg)
		assert.Equal(t, "json", cfg.Format)
	})

	t.Run("TestLogfileOption", func(t *testing.T) {
		cfg := defaultConfig()
		opt := WithLogFile("dingo.log")
		opt(cfg)
		assert.Equal(t, "dingo.log", cfg.LogFile)
	})
}

func TestLogLevels(t *testing.T) {
	core, recorded := observer.New(zap.DebugLevel)
	testLogger := &DingoLogger{
		zapLogger: zap.New(core),
		sugar:     zap.New(core).Sugar(),
	}

	originalLogger := globalLogger
	globalLogger = testLogger
	defer func() { globalLogger = originalLogger }()

	// should log at all levels
	t.Run("TestLogLevel", func(t *testing.T) {
		Debug("debug")
		Info("info")
		Warn("warn")
		Error("error")

		entries := recorded.All()
		assert.Equal(t, 4, len(entries))

		levels := []zapcore.Level{
			zapcore.DebugLevel,
			zapcore.InfoLevel,
			zapcore.WarnLevel,
			zapcore.ErrorLevel,
		}

		for i, entry := range entries {
			assert.Equal(t, levels[i], entry.Level)
		}
	})
}

func TestPanic(t *testing.T) {
	t.Run("TestPanic", func(t *testing.T) {
		testLogger := zaptest.NewLogger(t)
		dingoLogger := &DingoLogger{
			zapLogger: testLogger,
			sugar:     testLogger.Sugar(),
		}

		originalLogger := globalLogger
		globalLogger = dingoLogger
		defer func() { globalLogger = originalLogger }()

		func() {
			defer func() {
				if r := recover(); r != nil {
					assert.Contains(t, r.(string), "panic test")
				}
			}()
			Panic("panic test")
		}()
	})
}
