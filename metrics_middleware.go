package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/Sid-80/Redis-go/store"
)

// Compile-time check that *MetricMiddleware implements store.Storer.
var _ store.Storer = (*MetricMiddleware)(nil)

// MetricMiddleware wraps any Storer and counts calls, misses and latency.
// The counters are protected by mu so it can be used from many goroutines.
type MetricMiddleware struct {
	inner store.Storer

	mu sync.Mutex

	getCalls  int
	setCalls  int
	delCalls  int
	lenCalls  int
	keysCalls int
	incrCalls int

	getMisses  int
	incrErrors int

	totalGetLatency  time.Duration
	totalSetLatency  time.Duration
	totalIncrLatency time.Duration
}

func NewMetricMiddleware(inner store.Storer) *MetricMiddleware {
	return &MetricMiddleware{inner: inner}
}

// The call to the inner store happens outside the lock, so a slow store
// does not stop other goroutines from recording their metrics.

func (m *MetricMiddleware) Get(key string) (string, error) {
	start := time.Now()
	val, err := m.inner.Get(key)
	elapsed := time.Since(start)

	m.mu.Lock()
	m.getCalls++
	if err != nil {
		m.getMisses++
	}
	m.totalGetLatency += elapsed
	m.mu.Unlock()

	return val, err
}

func (m *MetricMiddleware) Set(key string, value string) error {
	start := time.Now()
	err := m.inner.Set(key, value)
	elapsed := time.Since(start)

	m.mu.Lock()
	m.setCalls++
	m.totalSetLatency += elapsed
	m.mu.Unlock()

	return err
}

func (m *MetricMiddleware) Del(key string) {
	m.inner.Del(key)

	m.mu.Lock()
	m.delCalls++
	m.mu.Unlock()
}

func (m *MetricMiddleware) Len() int {
	got := m.inner.Len()

	m.mu.Lock()
	m.lenCalls++
	m.mu.Unlock()

	return got
}

func (m *MetricMiddleware) Keys() []string {
	got := m.inner.Keys()

	m.mu.Lock()
	m.keysCalls++
	m.mu.Unlock()

	return got
}

func (m *MetricMiddleware) INCR(key string) (int, error) {
	start := time.Now()
	got, err := m.inner.INCR(key)
	elapsed := time.Since(start)

	m.mu.Lock()
	m.incrCalls++
	if err != nil {
		m.incrErrors++
	}
	m.totalIncrLatency += elapsed
	m.mu.Unlock()

	return got, err
}

func (m *MetricMiddleware) Report() {
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Printf("GET  calls: %d (missed %d)\n", m.getCalls, m.getMisses)
	fmt.Printf("SET  calls: %d\n", m.setCalls)
	fmt.Printf("INCR calls: %d (errors %d)\n", m.incrCalls, m.incrErrors)
	fmt.Printf("DEL  calls: %d\n", m.delCalls)
	fmt.Printf("LEN  calls: %d\n", m.lenCalls)
	fmt.Printf("KEYS calls: %d\n", m.keysCalls)

	printAvg("GET", m.totalGetLatency, m.getCalls)
	printAvg("SET", m.totalSetLatency, m.setCalls)
	printAvg("INCR", m.totalIncrLatency, m.incrCalls)
}

func printAvg(op string, total time.Duration, calls int) {
	if calls > 0 {
		fmt.Printf("avg %s latency: %v\n", op, total/time.Duration(calls))
	}
}
