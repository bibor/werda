Werda
=====

Notifies you if someone logged into your server via SSH. It parses the output of
`journalctl` and pushes a message to a [gotify](https://gotify.net/] server if a
new login via ssh occured.

Dpendencies
-----------

There are no build dependencies besides the go standard library. The only
runtime dependencies are `journalctl` and the libc, it was linkend to.
Currently the project only supports gotify as a backend. To use the gotify
backend you will need a [gotify](https://gotify.net/) server and gotify
clients to receive the alert messages.


Instructions
------------

### Build

```
    go build werda.go
```

### run

```
    export GOTIFYSERVER=<https://your.server>
    export GOTIFYTOKEN=<yourtoken>
    ./werda
```


Concept
----

Goal of the project is to create a small daemon to drop on servers to monitor
ssh login activity and forward those events.

The events should be able to sent to moile devices in near-real time.


### Implementation

To archive the goals the daemon monitors the syslog through `journalctl` and
waits for ssh login events. This event are parsed, pretified and than forwarded
to a third-party service.

Currently only *gotify* is supported as a message delivery backend, but I want
to add an IRC backend in the future. More complex protocols are not planned,
because it would lead either to an error prone implementation of the protocol in
this project or pull in huge dependencies.
currently the Daemon you have to drop are just a couple lines of code, with no
dpendencies, besides the standard library, and I think this is neat.

### Example Message from journal

~~~
$ journalctl -f -u sshd -o json
[...]
{
    "_SYSTEMD_CGROUP":"/system.slice/sshd.service",
    "_SYSTEMD_SLICE":"system.slice",
    "_SYSTEMD_UNIT":"sshd.service",
    "_CAP_EFFECTIVE":"1ffffffffff",
    "_HOSTNAME":"[REDACTED",
    "MESSAGE":"Accepted publickey for user from 192.168.0.106 port 45832 ssh2: ED25519 SHA256:[REDACTED]",
    "_GID":"0",
    "_CMDLINE":"sshd: [accepted]",
    "_EXE":"/nix/store/d8kdwl5k901l6yg67xjaz8vb69p1gnky-openssh-8.6p1/bin/sshd",
    "PRIORITY":"6",
    "SYSLOG_PID":"428863",
    "__MONOTONIC_TIMESTAMP":"108593[...]",
    "__CURSOR":"s=55bd17b5[...];i=75609;b=8454fda[...];m=fcd70e[...];t=5ce5659[...];x=d7d522fdw[...]",
    "__REALTIME_TIMESTAMP":"163424[...]",
    "_COMM":"sshd",
    "_PID":"428863",
    "_BOOT_ID":"8454fda97[...]",
    "_SELINUX_CONTEXT":"kernel",
    "_SOURCE_REALTIME_TIMESTAMP":"16342451[...]",
    "_UID":"0",
    "_TRANSPORT":"syslog",
    "SYSLOG_TIMESTAMP":"Oct 14 20:59:09 ",
    "SYSLOG_FACILITY":"4",
    "_SYSTEMD_INVOCATION_ID":"eeef868b61064d0481a408e[...]",
    "_MACHINE_ID":"[REDACTED]",
    "SYSLOG_IDENTIFIER":"sshd"
}
[...]
~~~
