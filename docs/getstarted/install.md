# Quick start

### With docker
Run mongodb server

use the [db init script](https://github.com/apache/servicecomb-kie/blob/master/deployments/db.js)

```shell script
sudo docker run --name mongo -d \
    -e "MONGO_INITDB_DATABASE=servicecomb" \
    -e "MONGO_INITDB_ROOT_USERNAME=kie" \
    -e "MONGO_INITDB_ROOT_PASSWORD=123" \
    -p 27017:27017 \
    -v ./deployments/db.js:/docker-entrypoint-initdb.d/db.js:ro \
    mongo:4.0
```
```shell script
export MONGO_IP=`sudo docker inspect --format '{{ .NetworkSettings.IPAddress }}' mongo`
```
Run kie server
```shell script
sudo docker run --name kie-server -d \
    -e "MONGODB_ADDR=${MONGO_IP}:27017" \
    -e "MONGODB_USER=kie" \
    -e "MONGODB_PWD=123" \
    -p 30110:30110 \
    servicecomb/kie
```


### Run locally with Docker compose

```bash
git clone git@github.com:apache/servicecomb-kie.git
cd servicecomb-kie/deployments/docker
sudo docker-compose up
```
it will launch 3 components 
- mongodb: 127.0.0.1:27017
- mongodb UI:http://127.0.0.1:8081
- servicecomb-kie: http://127.0.0.1:30110

### Run on kubernetes

```bash
kubectl apply -f https://raw.githubusercontent.com/apache/servicecomb-kie/master/deployments/kuberneetes/
```

it will launch 3 components, you can access them both in kubernetes and out of kubernetes.
out of kubernetes:
- mongodb: ${ANY_NODE_HOST}:30112
- mongodb UI:http://${ANY_NODE_HOST}:30111
- servicecomb-kie: http://${ANY_NODE_HOST}:30110
in kubernetes:
- mongodb: servicecomb-kie-nodeport:27017
- mongodb UI: servicecomb-kie-nodeport:8081
- servicecomb-kie: servicecomb-kie-nodeport:30110

