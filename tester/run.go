package main

import (
	"fmt"
	"time"
	
)

func main() {
	t := time.Now()
	fmt.Println(t.Month())
	fmt.Println(t.Day())
	fmt.Println(t.Year())

	state, err := database.NewStateFromDisk()
	if err != nil {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
	}
	defer state.Close()

}
