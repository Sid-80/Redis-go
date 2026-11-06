package main

import "fmt"

func main() {
	st := NewStore()
	st.Set("Key1", "Val1")

	fmt.Println(st.Get("Key1"))

	st.Del("Key1")
	fmt.Println(st.Get("Key1"))

}
