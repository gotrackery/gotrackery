# GoTrackery
![LintAndTest](https://github.com/gotrackery/gotrackery/actions/workflows/vet.yml/badge.svg)

- 🔭 I’m currently working on ...
- 👯 I’m looking to collaborate on ...

GoTrackery provides a server for obtaining geo-location from GPS trackers of their various protocols.
The extracted geo-location data is converted into a single format and allows subscribers to process them in a single way.
GoTrackery is CLI application made with Cobra.

GoTrackery has next features:
- Decode various telematic protocols;
- Handling various implementations of protocols (some vendors has not accurate implementation);
- Converting decoded data to unified data structure;
- Passing unified data structure to event subscrivers to support use for any purpose (i.e. saving into database, send message into message broker, calculate geofence events);
- Providing simple interfaces to implement new protocols.

Now implemented:
- Universal TCP server for tcp protocols;
- Universal TCP data replayer;
- Storing data into postgres example db;
- WialonsIPS protocol (partially - not all message types, no encoder);
- EGTS protocol (partially - not all message types);

## How to
### Build
Use `make` command to build executable:
- Linux: `make linux`
- Windows: `make win`
- MacOs Intel: `make mac`

### Run
Use `go run ./ --help` to get info about flags and features.
- Run server with wialonips protocol: `go run ./ tcp -p wialonips -a <address:port>`
- Run player with previously recorded data. See how to do it [here](./doc/tcpdump.md).

### Store data
For now implemented only posgtresql storage.
- Run `make db-up POSTGRES=<connection string>` to create tables and stored proc in target postgre database;
- Create yaml config:
```yaml
logging:
  level: debug
  console: true
tcp:
  address: :5000
  proto: egts
  socket-reuse-port: true
consumers:
  sample-db:
    uri:  <postgre connection string>
```
- Run server `go run ./ tcp --config <path to config file>`

### Docker
It is possible to use prebuild [docker image](https://hub.docker.com/r/gotrackery/gotrackery).

Also you can find `docker-compose.yml` example in `examples` folder to run along with Consul.

## The Future
### Shorterm Roadmap
- open telemetry (for server)
- replace generic message with protobuf message
- add mqtt event consumer with protobuf message
- protocol - wialon retranslator
- protocol - h02

Future plans for the implementation of other protocols see https://github.com/gotrackery/protocol

### Midterm Roadmap
- move database store to mqtt subscriber
- more protocols (udp, mqtt, http)
- implement encoders (wialon ips)
- retranslators as mqtt subscriber

### Longterm Roadmap
- Frontend
- Retranslator filter and subscribers
- Geofences https://github.com/peterstace/simplefeatures
- Event notificator

<!--
**gotrackery/gotrackery** is a ✨ _special_ ✨ repository because its `README.md` (this file) appears on your GitHub profile.

Here are some ideas to get you started:
- 🤔 I’m looking for help with ...
- 💬 Ask me about ...
- 📫 How to reach me: ...
-->
