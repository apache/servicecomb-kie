# Preparing for Development

### Setting Go
follow the official website to install

### Build
```shell
cd examples/dev
go build github.com/apache/servicecomb-kie/cmd/kieserver
```

### Setting up Config File
the most important thing is set the persistence storage kind, you can set it to "embedded_etcd", 
so that you don't need to set up an etcd or mongodb 
```shell
vim examples/dev/kie-conf.yaml
```
```yaml
db:
  # kind can be mongo, etcd, embedded_etcd
  kind: embedded_etcd
  # uri is the db endpoints list
  #   kind=mongo, then is the mongodb cluster's uri, e.g. mongodb://127.0.0.1:27017/kie
  #   kind=etcd, then is the  remote etcd server's advertise-client-urls, e.g. http://127.0.0.1:2379
  #   kind=embedded_etcd, then is the embedded etcd server's advertise-peer-urls, e.g. default=http://127.0.0.1:2380
  #uri: mongodb://kie:123@127.0.0.1:27017/kie
```
### Setting up go chassis config file
kie is developed base on go chassis and uses most of its features, go chassis need config file to launch itself

```shell
cd examples/dev/conf
```
move conf folder to same level with "kieserver" binary, if you move binary to  "examples/dev", 
you don't need to move or change those files
### Run
Now you can launch your config server
```shell
./kieserver --config kie-conf.yaml
```