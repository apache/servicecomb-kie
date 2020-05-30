# Quick start

### With docker
Run mongodb server

use the [db init script](https://github.com/apache/servicecomb-kie/blob/master/deployments/db.js)

```shell script
sudo docker run --name mongo -d \
    -e "MONGO_INITDB_DATABASE=kie" \
    -e "MONGO_INITDB_ROOT_USERNAME=root" \
    -e "MONGO_INITDB_ROOT_PASSWORD=root" \
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
    -e "MONGODB_USER=root" \
    -e "MONGODB_PWD=root" \
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

