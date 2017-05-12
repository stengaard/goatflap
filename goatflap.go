// Command goatflap
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"

	"os/signal"
	"syscall"
	"time"
)

type filer interface {
	File() (*os.File, error)
}

var msg = `Usage of %s

	   %s [opts] <command> [cmd args]

%s will run <command> with the supplied arguments.

Before starting <command> %s will start listening on the port supplied by -p
and pass the socket file descriptor to the child(ren).

%s can run several child processes at once and they will then share the listening
socket.

Parameters:

`

func init() {
	flag.Usage = func() {

		me := path.Base(os.Args[0])
		// Wauv. So needy.
		fmt.Fprintf(os.Stderr, msg, me, me, me, me, me)
		flag.PrintDefaults()
		os.Exit(1)
	}
}
func main() {

	var (
		addr    = flag.String("addr", ":5000", "Address to listen on")
		nt      = flag.String("net", "tcp", "Network type to listen on (tcp or unix)")
		runners = flag.Int("c", 1, "Number of concurrent children to start")
	)

	flag.Parse()

	if flag.NArg() == 0 {
		return
	}

	args := flag.Args()
	bin, args := args[0], args[1:]

	ln, err := net.Listen(*nt, *addr)
	if err != nil {
		log.Fatalf("goatflap: could not bind to port %v", err)
	}

	filer, ok := ln.(filer)
	if !ok {
		log.Fatalf("goatflap: listener type %T cannot be converted into a file descriptor", ln)
	}

	fd, err := filer.File()
	if err != nil {
		log.Fatalf("goatflap: could not get file descriptor for listening socket: %v", err)
	}

	procs := []*exec.Cmd{}
	exit := false
	procNum := 0
	// signal to kill one child process
	procDied := make(chan *exec.Cmd)
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGINT)
	var killer <-chan time.Time
	var toKill int

	for {

		// ensure N copies of the program is running
		for len(procs) < *runners && !exit {
			procNum++
			cmd := exec.Command(bin, args...)
			cmd.ExtraFiles = []*os.File{fd}
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("GOAT_PROC=%d", procNum),
				// the fd for tcp socket is 3. 0, 1 and 2 is stdin, stdout and stderr.
				// TODO(stengaard): make this configurable? why and how?
				fmt.Sprintf("LISTEN_FD=%d", 3),
			)

			err := cmd.Start()
			if err != nil {
				log.Fatalf("goatflap: could not run %q: %v", strings.Join(flag.Args(), " "), err)
			}

			go func(n int) {
				err := cmd.Wait()
				if err != nil {
					log.Printf("%d: %v", n, err)
				} else {
					log.Printf("%d: exited with 0 code", n)
				}
				procDied <- cmd
			}(procNum)

			procs = append(procs, cmd)

		}

		if exit && len(procs) == 0 {
			log.Println("all done.")
			return
		}

		// kills a process at a time in lock-step, this gives us graceful restarts.
		select {
		case sig := <-sigs:
			if exit {
				continue
			}

			if sig == syscall.SIGINT {
				log.Println("draining workers...")
				exit = true
			}

			killer = time.After(1 * time.Millisecond)
			// kill all the processes we have running
			toKill = len(procs)

		case <-killer:
			if len(procs) > 0 {
				procs[0].Process.Kill()
			}
			killer = nil

		case cmd := <-procDied:
			// cmd has died, so remove it from our state.
			// This also leaves cmds one element shorter - which in turn makes us
			// spawn more

			// if we haven't killed as many processes yet as was ordered - kill one more
			if toKill > 0 {
				d := 100 * time.Millisecond
				if exit {
					d = 10 * time.Millisecond
				}
				killer = time.After(d)
				toKill--
			}

			for i := 0; i < len(procs); i++ {
				if procs[i] == cmd {
					copy(procs[i:], procs[i+1:])
					procs[len(procs)-1] = nil // in order to let GC collect this element
					procs = procs[:len(procs)-1]
				}
			}
		}
	}
}
