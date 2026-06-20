package fromscratch

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func reUsePort(network, add string) (net.Listener, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var sockErr error
			err := c.Control(func(fd uintptr) {
				sockErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
				if sockErr != nil {
					fmt.Println("Error while getting the socket")
					return
				}
				sockErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)

			})
			if err != nil {
				fmt.Println("error in COntrol func")
				return err
			}
			return sockErr
		},
	}
	return lc.Listen(context.Background(), network, add)
}
func Run() {
	pid := os.Getpid()
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("HEllo from route PID %d", pid)))
	})

	ln1, err := reUsePort("tcp", ":8080")
	if err != nil {
		log.Fatal("failed to get listener ")
	}
	fmt.Println("starting  go server with PID", pid)
	http.Serve(ln1, nil)
}
