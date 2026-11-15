package main

import (
	"log"
	"os"

	"github.com/Sid-80/Redis-go/store"
)

// Compile-time check that *LoggingMiddleware implements store.Storer.
var _ store.Storer = (*LoggingMiddleware)(nil)

// LoggingMiddleware wraps any Storer and logs every call before passing it on.
// log.Logger is safe for concurrent use, so no extra locking is needed here.
type LoggingMiddleware struct {
	inner  store.Storer
	logger *log.Logger
}

func NewLoggingMiddleware(inner store.Storer) *LoggingMiddleware {
	return &LoggingMiddleware{
		inner:  inner,
		logger: log.New(os.Stdout, "[log] ", log.Ltime|log.Lmicroseconds),
	}
}

func (l *LoggingMiddleware) Get(key string) (string, error) {
	val, err := l.inner.Get(key)
	if err != nil {
		l.logger.Printf("GET %q -> miss (%v)", key, err)
	} else {
		l.logger.Printf("GET %q -> hit", key)
	}
	return val, err
}

func (l *LoggingMiddleware) Set(key string, value string) error {
	err := l.inner.Set(key, value)
	if err != nil {
		l.logger.Printf("SET %q -> error (%v)", key, err)
	} else {
		l.logger.Printf("SET %q", key)
	}
	return err
}

func (l *LoggingMiddleware) Del(key string) {
	l.logger.Printf("DEL %q", key)
	l.inner.Del(key)
}

func (l *LoggingMiddleware) Len() int {
	got := l.inner.Len()
	l.logger.Printf("LEN -> %d", got)
	return got
}

func (l *LoggingMiddleware) Keys() []string {
	got := l.inner.Keys()
	l.logger.Printf("KEYS -> %v", got)
	return got
}

func (l *LoggingMiddleware) INCR(key string) (int, error) {
	got, err := l.inner.INCR(key)
	if err != nil {
		l.logger.Printf("INCR %q -> error (%v)", key, err)
	} else {
		l.logger.Printf("INCR %q -> %d", key, got)
	}
	return got, err
}
