package main

import (
	"log"
	"sync"

	"github.com/Sid-80/Redis-go/store"
)

type Command struct {
	Op    string
	Key   string
	Value string
}

func dispatch(s store.Storer, c Command) {
	switch c.Op {
	case "SET":
		log.Printf("SET Key: %s", c.Key)
		s.Set(c.Key, c.Value)
	case "GET":
		s.Get(c.Key)
	case "INCR":
		s.INCR(c.Key)
	}
}

func restartsOnBoot(s store.Storer, cmds []Command) {
	var wg sync.WaitGroup
	for _, c := range cmds {
		wg.Add(1)
		go func() {
			dispatch(s, c)
			wg.Done()
		}()
	}

	wg.Wait()
}
