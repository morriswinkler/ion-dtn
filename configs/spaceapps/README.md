# Setup
Run

 - ionstart -I tcp.1.rc

which calls `ionadmin`, `bpadmin` and `ipadmin` to perform the setup.

# Testing

> :whale: The assumption is made that these cases are tested with the
> [Docker compose](docker-compose.yml) setup which may be started with
> `docker-compose up`.

Start a shell session inside node 1, using

```
docker-compose exec node-1 /bin/bash
```

and use [`bpecho`](bp/test/bping.c) as follows:


```
bpecho ipn:1.1
```

to initiate a receiver for bundles on endpoint `ipn:1.1`.

> NOTE: Without first defining a listener (e.g.: `bpecho`) any connection
> attempt (e.g.: `bping`) will fail.


In order to confirm the existence of a connection, start a shell session in
node 2 using

```
docker-compose exec node-2 /bin/bash
```

and issue pings with [`bping`](bp/test/bping.c) by running

```
bping -c 10 -v ipn:2.1 ipn:1.1
```

which will ping from `ipn:2.1` to `ipn1.1`.

> NOTE: The source EID can be any valid EID that has been associated to the
> current node, so from node 2 we can ping from EID's `ipn:2.1`, `ipn:2.2` and
> `ipn:2.3`, but not from any of the `ipn:1.*` EID's.
