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

Inspiration is from [catflap](https://github.com/passcod/catflap), but in Go. And with Goats. Goats also need to get back in.

Installation
------------

        go get -u github.com/stengaard/goatflap

Usage
------

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


Notes
-----
Do not use this as loadbalancing. -c larger than 2 will give you little benefit. What this _does_ offer you is graceful reloads, not load balancing.

Running `goatflap -c 10 examples/http_server` and then shooting it with [rakyll/hey](https://github.com/rakyll/hey) a few times - this is the result (`hey -n 100000 http://localhost:5000/`)

        2017/05/12 12:02:29 goat proc 1 0
        2017/05/12 12:02:29 goat proc 2 2400
        2017/05/12 12:02:29 goat proc 3 0
        2017/05/12 12:02:29 goat proc 4 8800
        2017/05/12 12:02:29 goat proc 6 46400
        2017/05/12 12:02:29 goat proc 5 600
        2017/05/12 12:02:29 goat proc 7 2400
        2017/05/12 12:02:29 goat proc 8 6000
        2017/05/12 12:02:29 goat proc 9 12000
        2017/05/12 12:02:29 goat proc 10 41400

Notice the large deviation. Two child processes did more than 80% of the work and two child
processes did not pick up a single HTTP request.