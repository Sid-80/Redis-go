package main

import (
	"fmt"
	"time"

	"github.com/Sid-80/Redis-go/store"
)

type MetricMiddleware struct {
	inner store.Storer

	getCalls int
	setCalls int
	delCalls int
	lenCalls int

	getMisses int

	totalGetLatency time.Duration
	totalSetLatency time.Duration
}

func NewMetricMiddleware(inner store.Storer) *MetricMiddleware {
	return &MetricMiddleware{
		inner:     inner,
		getCalls:  0,
		setCalls:  0,
		delCalls:  0,
		lenCalls:  0,
		getMisses: 0,
	}
}

func (m *MetricMiddleware) Get(key string) (string, error) {
	start := time.Now()
	m.getCalls++
	val, err := m.inner.Get(key)

	if err != nil {
		m.getMisses++
	}

	duration := time.Since(start)
	m.totalGetLatency += duration

	return val, err
}

func (m *MetricMiddleware) Set(key string, value string) error {
	start := time.Now()

	m.setCalls++

	err := m.inner.Set(key, value)

	duration := time.Since(start)
	m.totalSetLatency += duration

	return err
}

func (m *MetricMiddleware) Del(key string) {
	m.delCalls++
	m.inner.Del(key)
}

func (m *MetricMiddleware) Len() int {
	m.lenCalls++
	return m.inner.Len()
}

func (m *MetricMiddleware) Keys() []string {
	return m.inner.Keys()
}

func (m *MetricMiddleware) Report() {
	fmt.Printf("GET calls: %d. (missed %d)\n", m.getCalls, m.getMisses)
	fmt.Printf("SET calls: %d.\n", m.setCalls)
	fmt.Printf("DEL calls: %d.\n", m.delCalls)
	fmt.Printf("LEN calls: %d.\n", m.lenCalls)

	if m.getCalls > 0 {
		avg := m.totalGetLatency / time.Duration(m.getCalls)
		fmt.Printf("avg GET latency: %v.\n", avg)
	}

	if m.setCalls > 0 {
		avg := m.totalSetLatency / time.Duration(m.setCalls)
		fmt.Printf("avg SET latency: %v.\n", avg)
	}

}
