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
>*(optional, bool)*  if sslEnabled is true, you must give a ca file


### Example
```yaml
db:
  uri: mongodb://kie:123@127.0.0.1:27017
  poolSize: 10
  timeout:  5m
  sslEnabled: true
  rootCAFile: /opt/kie/ca.crt
```


