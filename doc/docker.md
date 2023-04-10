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

## Run player