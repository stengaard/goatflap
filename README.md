goatflap
---------
Start a listening socket and pass the socket to a child process.

Example:

        # This starts a HTTP server on :5000 and distributes calls to proc 1 and proc 2
        $ goatflap -c 2 examples/http_server
        2017/05/12 11:55:35 goat proc 1 0
        2017/05/12 11:55:35 goat proc 2 0
        2017/05/12 11:55:36 goat proc 2 0
        2017/05/12 11:55:37 goat proc 1 0
        2017/05/12 11:55:38 goat proc 1 0
        2017/05/12 11:55:38 goat proc 2 0
        ^C2017/05/12 11:55:38 draining workers...
        2017/05/12 11:55:38 1: signal: interrupt
        2017/05/12 11:55:38 2: signal: interrupt
        2017/05/12 11:55:38 all done.

Note: Do not use this as loadbalancing. -c larger than 2 will give you little benefit. What this _does_ offer you is graceful reloads, not load balancing.

Inspiration is from [catflap](https://github.com/passcod/catflap), but in Go. And with Goats. Goats also need to get back in.

        Usage of goatflap

            goatflap [opts] <command> [cmd args]

        goatflap will run <command> with the supplied arguments.

        Before starting <command> goatflap will start listening on the port supplied by -p
        and pass the socket file descriptor to the child(ren).

        goatflap can run several child processes at once and they will then share the listening
        socket.

        goatflap reacts to SIGUSR1 with a reload of all running child processes. SIGINT (Ctrl-C)
        stops all child processes and exits.

        Parameters:

        -addr string
                Address to listen on (default ":5000")
        -c int
                Number of concurrent children to start (default 1)
        -net string
                Network type to listen on (tcp or unix) (default "tcp")