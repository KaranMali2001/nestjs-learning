# Two Processes on the Same Port (`SO_REUSEPORT`)

A small lab that demonstrates how multiple, independent processes can bind and
serve on the **same TCP port** using the `SO_REUSEPORT` socket option — and how
the same code behaves **differently on macOS vs Linux**.

## The idea

Normally a port can only be bound by one socket. Try to `bind()` a second socket
to the same port and the kernel rejects it with `address already in use`
(`EADDRINUSE`). This is a **per-socket** rule, not a per-process one — two
goroutines in the *same* process hit it just as two separate processes do.

`SO_REUSEPORT` is the opt-in that relaxes the rule: every socket that sets it
*before* `bind()` is allowed to share the port, and the kernel gives each its own
accept queue.

What you actually get depends on the OS:

| OS | Two processes bind `:8080`? | Connection distribution |
|----|------------------------------|--------------------------|
| **Linux** | ✅ | Load-balanced across processes via a 4-tuple hash (≈50/50) |
| **macOS / BSD** | ✅ | **Last process to bind wins** — it takes essentially all connections |

## Layout

```
.
├── main.go                     # entry point → runs the from-scratch server
├── from-scratch/               # SO_REUSEPORT from scratch (no library)
│   └── own.go                  #   reUsePort() via net.ListenConfig.Control + setsockopt
├── using-library/              # two servers in ONE process (two goroutines)
│   └── artificial.go           #   uses the go-reuseport library
├── go-reuseport/               # vendored copy of libp2p/go-reuseport (for reading)
├── Dockerfile                  # golang:1.26-alpine + curl
└── docker-compose.yml          # one "lab" container to run the Linux demo
```

The whole trick lives in `from-scratch/own.go`:

```go
lc := net.ListenConfig{
    Control: func(network, address string, c syscall.RawConn) error {
        var sockErr error
        c.Control(func(fd uintptr) {
            // must be set BEFORE bind() — that's why we use the Control hook
            sockErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
            if sockErr != nil { return }
            sockErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
        })
        return sockErr
    },
}
return lc.Listen(context.Background(), network, add)
```

Each server tags its `/hello` response with its own **PID** so you can see which
process answered a given request.

## Run it — macOS (last-binder-wins)

Open **two terminals**, in each:

```bash
go run .
```

Then hammer the endpoint from a third terminal:

```bash
for i in $(seq 1 20); do curl -s localhost:8080/hello; echo; done
```

You'll see the **same PID** every time — macOS hands all connections to whichever
process bound last. (If you only see one PID and the other terminal errored, make
sure you ran `go run .` — not `go run main.go` — from this directory.)

## Run it — Linux (load-balanced) via Docker

This is where you see the kernel actually spread the load. One container, two
processes inside it, curling from **inside** the container:

```bash
docker compose run --rm -T lab sh -c '
  go build -o /tmp/server .
  /tmp/server & /tmp/server &
  sleep 2
  for i in $(seq 1 200); do curl -s localhost:8080/hello; echo; done | sort | uniq -c'
```

Expected output (PIDs and exact split vary per run):

```
starting  go server with PID 1962
starting  go server with PID 1963
    102 HEllo from route PID 1962
     98 HEllo from route PID 1963
```

Both PIDs serve roughly half the traffic. The split is **statistical** (a hash of
each connection's 4-tuple), so it wobbles around 50/50 — it is *not* perfect
round-robin.

> Curl from **inside** the container, not from your Mac against the mapped port.
> Docker Desktop's host→container proxy can flatten the source 4-tuple and hide
> the distribution.

## Notes & gotchas

- **`http.Serve` vs `http.ListenAndServe`** — `ListenAndServe` creates its *own*
  plain listener and ignores any socket you built. To serve on a `SO_REUSEPORT`
  listener you **must** use `http.Serve(ln, handler)`.
- **The option must precede `bind()`** — hence the `Control` hook, which Go calls
  after the socket is created but before it's bound.
- **The library is ~3 lines** — `github.com/libp2p/go-reuseport` does exactly the
  two `setsockopt` calls above; the kernel does all the real work.
- **Not for zero-downtime deploys** — for graceful upgrades, `SO_REUSEPORT` has a
  race where connections in the old process's accept queue get orphaned on exit.
  Production setups (NGINX, HAProxy, Cloudflare's `tableflip`) pass the listener
  file descriptor to the new process instead.
