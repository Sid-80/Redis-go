package main

import (
	"fmt"

	"github.com/Sid-80/Redis-go/store/kv"
)

func main() {

	s := kv.NewStore(20)

	s.Set("pageViews", "0")

	const hits = 1000
	cmds := make([]Command, hits)

	for i := range cmds {
		cmds[i] = Command{Op: "INCR", Key: "pageViews"}
	}

	restartsOnBoot(s, cmds)

	got, _ := s.Get("pageViews")

	fmt.Println("Got: ", got)
	fmt.Println("Req: ", hits)
}
