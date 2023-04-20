### Starting the Scanner
Scan can be started by either providing ips or subscribing to event queue.

You can run it by using  `docker compose up` or compiling binary `make scanner` 


### Settings

If started with cache, the scanner will refuse to scan twice the same ip within 12 hours.

If started with db, the scanner will automatically insert results into db.

Ratelimit is enabled by default. You can tune it down by using `burst` and `interval_in_ms` flags.

You can scan limit scan to range ports by providing `port_start` and `port_end` flags.

For more detailed information and options use `--help` command


