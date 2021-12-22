# Storage 
you can use mongo db as kie server storage to save configuration

### Options
**uri**
>*(required, string)* db uri


**timeout**
>*(optional, string)* db operation timeout 


**sslEnabled**
>*(optional, bool)*  enable TLS communication to mongodb server

**rootCAFile**
>*(optional, string)*  if sslEnabled is true, you must give a ca file

**verifyPeer**
>*(optional, bool)*  if verifyPeer is true, kie will verify database server's certificate, otherwise not


### Example
```yaml
db:
  uri: mongodb://kie:123@127.0.0.1:27017/servicecomb
  poolSize: 10
  timeout:  5s
  sslEnabled: true
  rootCAFile: /opt/kie/ca.crt
  verifyPeer: true
```


