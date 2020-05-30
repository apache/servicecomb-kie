# After running
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

response is 
```json
{
    "id": "b01ad993-2bad-4468-8a3c-5aa7ad54afea",
    "key": "timeout",
    "value": "2s",
    "value_type": "text",
    "create_revision": 2,
    "update_revision": 2,
    "status": "disabled",
    "create_time": 1590802244,
    "update_time": 1590802244,
    "labels": {
        "service": "order"
    }
}
```

then get config list
```shell script
curl -X GET http://127.0.0.1:30110/v1/default/kie/kv 
```
response is 
```json
{
 "total": 1,
 "data": [
  {
   "id": "b01ad993-2bad-4468-8a3c-5aa7ad54afea",
   "key": "timeout",
   "value": "2s",
   "value_type": "text",
   "create_revision": 2,
   "update_revision": 2,
   "status": "disabled",
   "create_time": 1590802244,
   "update_time": 1590802244,
   "labels": {
    "service": "order"
   }
  }
 ]
}
```


Check open API doc
- the api doc mounted to http://127.0.0.1:30110/apidocs.json 
- or see https://github.com/apache/servicecomb-kie/blob/master/docs/api.yaml
