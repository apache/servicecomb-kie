# Quick start

### With docker
Run mongodb server

write a script to create a user 
```shell script
cat <<EOM > db.js
db.createUser(
    {
        user: "root",
        pwd: "root",
        roles:[
            {
                role: "readWrite",
                db:   "kie"
            }
        ]
    }
);
EOM
```

```shell script
sudo docker run --name mongo -d \
    -e "MONGO_INITDB_DATABASE=kie" \
    -e "MONGO_INITDB_ROOT_USERNAME=root" \
    -e "MONGO_INITDB_ROOT_PASSWORD=root" \
    -p 27017:27017 \
    -v ${PWD}/db.js:/docker-entrypoint-initdb.d/db.js:ro \
    mongo:3.4
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

Put a key 
```shell script
curl -X PUT \
  http://127.0.0.1:30110/v1/default/kie/kv/ingressRule.http \
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
    "_id": "5d6f27c5a1b287c5074e4538",
    "label_id": "5d6f27c5a1b287c5074e4537",
    "key": "ingressRule.http",
    "value": "rule",
    "value_type": "text",
    "labels": {
        "app": "default"
    },
    "revision": 1
}
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


### Run locally without Docker

Download and run mongodb, 
see [MongoDB Community Edition Installation Tutorials](https://docs.mongodb.com/manual/installation/#mongodb-community-edition-installation-tutorials)

Write a script to create a user 
```shell script
cat <<EOM > native_db.js
db.createUser(
    {
        user: "root",
        pwd: "root",
        roles:[
            {
                role: "readWrite",
                db:   "kie"
            }
        ]
    }
);
EOM
```

Exec native_db.js
```shell script bash
mongo 127.0.0.1/kie native_db.js
```

Download the binary of kie, see https://apache.org/dyn/closer.cgi/servicecomb/servicecomb-kie/0.1.0/
Unzip  and Run it.
For Example:
```shell script bash
tar -xzvf apache-servicecomb-kie--${platform}-amd64.tar.gz
cd apache-servicecomb-kie--${platform}-amd64
./kie --config conf/kie-conf.yaml
```