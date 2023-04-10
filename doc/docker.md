# GoTrackery

## Run server tcp server with certain protcol
Flags:
```bash
docker run --rm --name gotrackery \
	-p 5001:5001/tcp \
	gotrackery/gotrackery tcp -p wialonips -a :5001
```
Consul:
```bash
docker run --rm --name gotrackery \
	-p 5001:5001/tcp \
	gotrackery/gotrackery tcp --consul <consul address> --consul-key <consul key>
```
Example of consul key:
```json
{
  "logging": {
    "level": "debug",
    "console": true
  },
  "tcp": {
    "address": ":5039",
    "proto": "wialonips",
    "socket-reuse-port": true
  }
}
```

## Run player

```bash
docker run --rm --name gotrackery \
        -v <local path with extracted payload>:/gotrackery/payload \
        gotrackery/gotrackery replay -p wialonips -a <address> -i /gotrackery/payload
```