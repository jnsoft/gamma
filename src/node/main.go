package main

import (
	"flag"
	"fmt"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "IP address to listen on")
	port := flag.Int("p", 8080, "Port to listen on")
	useSSL := flag.Bool("ssl", false, "Enable SSL")
	verbose := flag.Bool("v", false, "Enable verbose logging")
	flag.Parse()

	if *verbose {
		fmt.Printf("Listening on: %s:%d\n", *ip, *port)
		if *useSSL {
			fmt.Println("SSL is enabled.")
		}
	}
}
