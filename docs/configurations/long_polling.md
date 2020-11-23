# distribute

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
```json
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

# Long polling
*experimental*

Kie leverage gossip protocol to broad cast cluster events. if client use query parameter "?wait=5s" to poll key value,
this polling will become long polling and if there is key value change events, 
server will response key values to client.

## For example 

Put a key 
```shell script
curl -X POST \
  http://127.0.0.1:30110/v1/default/kie/kv/ \
  -H 'Content-Type: application/json' \
  -d '{
    "key": "timeout",
    "value": "2s",
    "labels": {
        "service": "order"
    }
}'
```

Then, we get the id from response.
```json
{
    "id": "ec4d8057-fc3e-4a78-b273-28e311187196",
    "key": "timeout",
    "value": "2s",
    "value_type": "text",
    "create_revision": 2,
    "update_revision": 2,
    "status": "disabled",
    "create_time": 1596620728,
    "update_time": 1596620728,
    "labels": {
        "service": "order"
    }
}
```

Open the another terminal test the long polling
```shell script
 curl -X GET http://127.0.0.1:30110/v1/default/kie/kv?wait=20s
```
After 20 seconds, return empty body and the status code is 304

Then modify the value
```shell script
 curl -X PUT http://127.0.0.1:30110/v1/default/kie/kv/ec4d8057-fc3e-4a78-b273-28e311187196  \
 -H 'Content-Type: application/json'   \
 -d '{
     "value": "8s"
  }'
```

Terminal immediately shows the result
```json
{
 "total": 1,
 "data": [
  {
   "id": "ec4d8057-fc3e-4a78-b273-28e311187196",
   "key": "timeout",
   "value": "8s",
   "value_type": "text",
   "create_revision": 2,
   "update_revision": 3,
   "status": "disabled",
   "create_time": 1596620728,
   "update_time": 1596713915,
   "labels": {
    "service": "order"
   }
  }
 ]
}
```

