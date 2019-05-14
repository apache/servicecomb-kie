# Intro
this guide will show you how to develop servicecomb-kie in your local machine. 

servicecomb-kie only depend on mongodb, so you have 2 choices
- setup a mongodb and give credential in kie-conf.yaml
- setup a simple mongodb alone with admin UI by docker compose 

in this guide, we will use mongodb launched by docker compose

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

# mongodb admin UI
http://127.0.0.1:8081/

#servicecomb-kie endpoint
http://127.0.0.1:30110/

# API document
the API doc will be generated under 
examples/dev/conf/servicecomb-kie/schema

you can copy it to https://editor.swagger.io/ to see full API document
