## tsddrain
A sink for TSD traffic using the "raw" TCP protocol.

```
  -b="localhost:4242": bind host
  -o="":               output file
  -f=2s:               flush interval
```

Output file (`-o`) defaults to "", which will drop the messages on the floor.
It can be set to "-" (which will write to STDOUT) or a filename.
The file output is buffered and flushed every flush interval (`-f`) seconds.

Commands accepted over the TCP port are:
* `put` - expects a a regular OpenTSDB datapoint (though no validation is performed).
* `version` - prints a small version string (expected by tools such as [TCollector](https://github.com/OpenTSDB/tcollector))
* `stats` - prints some basic statistics.
* `reset` - resets the stats counters.

### Stats
* `tsdsink.events` - number of "put" events received since startup
* `tsdsink.bytes` - bytes received in "put" events since startup
* `tsdsink.connections` - total of connections made since startup
* `tsdsink.open` - number of open TCP connections
* `tsdsink.unknown` - number of unknown commands

These stats can be quickly dumped into a running TSD for graphing:
```
while true; do echo stats | nc localhost 4242 | awk "{print \"put\", \$0, \"host=$HOSTNAME\"}"; sleep 5; done | nc tsdhost.local 4242
```

### Building
Either `go build tsddrain.go` to generate a compiled "tsddrain" executable, or `go run tsddrain.go [-options]` to compile and run without generating an artifact.
