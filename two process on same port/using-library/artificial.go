package usinglibrary

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"

	"github.com/libp2p/go-reuseport"
)

func Run() {

	var wg sync.WaitGroup

	wg.Go(func() {
		defer wg.Done()
		server1()
	})

	wg.Go(func() {
		defer wg.Done()

		server2()
	})

	wg.Wait()
	fmt.Println("FINISHING THE SERVER")
}
func server1() {

	mux1 := http.NewServeMux()
	mux1.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello Route from server 1"))
	})
	if !reuseport.Available() {
		fmt.Println("cant spawn two processes on same port on this os:", runtime.GOOS)
	}

	ln1, err := reuseport.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("failed to get listener ")
	}
	fmt.Println("starting first go server", os.Getpid())
	http.ListenAndServe(":8080", nil)
	http.Serve(ln1, mux1)
}

func server2() {
	mux2 := http.NewServeMux()
	mux2.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello Route from server 2"))
	})
	if !reuseport.Available() {
		fmt.Println("cant spawn two processes on same port on this os:", runtime.GOOS)
	}

	ln1, err := reuseport.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("failed to get listener ")
	}
	fmt.Println("starting second go server", os.Getpid())
	http.Serve(ln1, mux2)
}
