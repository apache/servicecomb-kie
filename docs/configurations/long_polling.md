# Long polling
*experimental*

Kie leverage gossip protocol to broad cast cluster events. if client use query parameter "?wait=5s" to poll key value,
this polling will become long polling and if there is key value change events, 
server will response key values to client.

kie must join to a cluster and listen to peer events

start first node
```shell script
./kie --name=kie0 --listen-peer-addr=10.1.1.11:5000
```

start another node
```shell script
./kie --name=kie1 --listen-peer-addr=10.1.1.12:5000 --peer-addr=10.1.1.11:5000
```

### event payload and trigger condition

condition: key value put or delete

payload:
```go
{
  "Key": "timeout",
  "Action": "put",
  "Labels": {
    "app": "default",
    "service": "order"
  },
  "DomainID": "default",
  "Project": "default"
}
```