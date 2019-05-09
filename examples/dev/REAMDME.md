# Intro
that is a simple example to run kie in you local machine

you only need to set up a mongodb and config credential 
and related info in kie-conf.yaml

you can setup a simple mongodb alone with admin UI by docker compose 

# Get started

1.Build 
```bash
cd examples/dev
go build github.com/apache/servicecomb-kie/cmd/kie
```

2.Run mongodb and servicecomb-kie
```bash
sudo docker-compose up
./kie --config kie-conf.yaml
```

3. check service API document
```bash
cd examples/dev/conf/servicecomb-kie/schema
```
you can copy it to https://editor.swagger.io/ to see full API document
