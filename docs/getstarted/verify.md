# After running
Put a key 
```shell script
curl -X PUT \
  http://127.0.0.1:30110/v1/default/kie/kv/ingressRule \
  -H 'Content-Type: application/json' \
  -d '{
	"value":"some rule",
	"type": "yaml",
	"labels":{"app":"default"}
}'
```

response is 
```json
{
 "id": "05529229-efc3-49ca-a765-05759b23ab28",
 "key": "ingressRule",
 "value": "some rule",
 "value_type": "text",
 "create_revision": 13,
 "update_revision": 13,
 "labels": {
  "app": "default"
 }
}
```

Check open API doc
- the api doc mounted to http://127.0.0.1:30110/apidocs.json 
- or see https://github.com/apache/servicecomb-kie/blob/master/docs/api.yaml
