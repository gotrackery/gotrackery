# Replaying data from devices and retranslators

## Collect tcp dump and extract payload

 - Install `tcpdump` and `tcpflow`
 - To collect tcp dump run `sudo tcpdump -w dump.pcap -i <network interface> port <port>`
 - To extract payload run `sudo tcpflow -r dump.pcap -o <path to store files with payload>`

As result there will be set of files in the out directory.

## Replaying extracted payload
 - Run `go run ./ replay -a <server address> -p wialonips -i <path to stored files with payload>`
