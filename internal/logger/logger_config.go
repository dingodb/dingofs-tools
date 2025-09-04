package logger

const (
	DEFAULT_LOG_FILE   = "dingo.log"
	DEFAULT_LOG_LEVEL  = "info"
	DEFAULT_LOG_FORMAT = "text"
)

type logConfig struct {
	LogFile    string
	LogLevel   string
	Format     string
	MaxSize    int // MB
	MaxBackups int
	MaxAge     int // days
	Compress   bool
	Stdout     bool // Whether to output to stdout simultaneously
}

type Option func(*logConfig)

func defaultConfig() *logConfig {
	return &logConfig{
		LogFile:    DEFAULT_LOG_FILE,
		LogLevel:   DEFAULT_LOG_LEVEL,
		Format:     DEFAULT_LOG_FORMAT,
		MaxSize:    1024,
		MaxBackups: 1,
		MaxAge:     1,
		Compress:   false,
		Stdout:     false,
	}
}

func WithLogFile(logFile string) Option {
	return func(c *logConfig) {
		c.LogFile = logFile
	}
}

func WithLogLevel(logLevel string) Option {
	return func(c *logConfig) {
		c.LogLevel = logLevel
	}
}

func WithFormat(format string) Option {
	return func(c *logConfig) {
		c.Format = format
	}
}

func WithMaxSize(maxSize int) Option {
	return func(c *logConfig) {
		c.MaxSize = maxSize
	}
}

func WithMaxBackups(maxBackups int) Option {
	return func(c *logConfig) {
		c.MaxBackups = maxBackups
	}
}

func WithMaxAge(maxAge int) Option {
	return func(c *logConfig) {
		c.MaxAge = maxAge
	}
}

func WithCompress(compress bool) Option {
	return func(c *logConfig) {
		c.Compress = compress
	}
}

func WithStdout(stdout bool) Option {
	return func(c *logConfig) {
		c.Stdout = stdout
	}
}
