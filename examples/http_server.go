package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var count struct {
	sync.Mutex
	num int
}

func goatListener() (net.Listener, error) {
	fd, err := strconv.Atoi(os.Getenv("LISTEN_FD"))
	if err != nil {
		return nil, err
	}

	f := os.NewFile(uintptr(fd), "goatlistener")
	return net.FileListener(f)
}

func main() {
	l, err := goatListener()
	if err != nil {
		log.Fatalf("http_server: could not create listener: %s", err)
	}

	s := http.Server{
		Handler: http.HandlerFunc(handler),
	}
	go func() {
		for range time.Tick(1 * time.Second) {
			count.Lock()
			log.Println("goat proc", os.Getenv("GOAT_PROC"), count.num)
			count.Unlock()
		}
	}()

	s.Serve(l)
}

func handler(rw http.ResponseWriter, req *http.Request) {
	count.Lock()
	defer count.Unlock()

	count.num++
	fmt.Fprintf(rw, "Hello from proc %s. count=%d\n", os.Getenv("GOAT_PROC"), count.num)
}
