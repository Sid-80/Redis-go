package main

import (
	"fmt"
	"sync"

	"github.com/Sid-80/Redis-go/store"
)

type Command struct {
	Op    string
	Key   string
	Value string
}

// dispatch runs a single command against s. Logging is left to LoggingMiddleware.
func dispatch(s store.Storer, c Command) error {
	switch c.Op {
	case "SET":
		return s.Set(c.Key, c.Value)
	case "GET":
		_, err := s.Get(c.Key)
		return err
	case "DEL":
		s.Del(c.Key)
		return nil
	case "INCR":
		_, err := s.INCR(c.Key)
		return err
	default:
		return fmt.Errorf("unknown op: %q", c.Op)
	}
}

// restartsOnBoot replays cmds concurrently and returns every error that occurred.
func restartsOnBoot(s store.Storer, cmds []Command) []error {
	var wg sync.WaitGroup
	errs := make(chan error, len(cmds))

	for _, c := range cmds {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := dispatch(s, c); err != nil {
				errs <- fmt.Errorf("%s %q: %w", c.Op, c.Key, err)
			}
		}()
	}

	// Close errs once every worker is done, so the range loop below can finish.
	go func() {
		wg.Wait()
		close(errs)
	}()

	var all []error
	for err := range errs {
		all = append(all, err)
	}
	return all
}
