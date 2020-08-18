# RBAC
if you enabled service center [RBAC](https://service-center.readthedocs.io/en/latest/user-guides/rbac.html).

you can choose to enable kie RBAC feature, after enable RBAC, all request to kie must be authenticated

### Configuration file
follow steps to enable service center [RBAC](https://service-center.readthedocs.io/en/latest/user-guides/rbac.html).

1.get public key file which is exactly same with service center public key file

2.edit kie-conf.yaml
```ini
db:
  uri: mongodb://kie:123@127.0.0.1:27017/kie
  type: mongodb
rbac:
  enabled: true
  rsaPublicKeyFile: ./examples/dev/public.key
```

To distribute your public key, you can use kubernetes config map to manage public key
### Generate a token 
token is the only credential to access rest API, before you access any API, you need to get a token from service center
```shell script
curl -X POST \
  http://127.0.0.1:30100/v4/token \
  -d '{"name":"root",
"password":"rootpwd"}'
```
will return a token, token will expired after 30m
```json
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTI4MzIxODUsInVzZXIiOiJyb290In0.G65mgb4eQ9hmCAuftVeVogN9lT_jNg7iIOF_EAyAhBU"}
```

### Authentication
For each request you must add token to http header:
```
Authorization: Bearer {token}
```
for example:
```shell script
curl -X GET \
  'http://127.0.0.1:30110/v1/default/kie/kv' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTI4OTQ1NTEsInVzZXIiOiJyb290In0.FfLOSvVmHT9qCZSe_6iPf4gNjbXLwCrkXxKHsdJoQ8w' 
```