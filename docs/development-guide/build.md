# Build

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

Build serivcecomb-kie binary

```shell script bash
cd ${servicecomb_root_path}
build/build_binary.sh
```

This will build 4 packages for 4 different platforms in ${servicecomb_root_path}/release/kie directory.
Choose the right one and unzip and run it.
For Example:
```shell script bash
cd ${servicecomb_root_path}/release/kie
tar -xzvf apache-servicecomb-kie--${platform}-amd64.tar.gz
cd apache-servicecomb-kie--${platform}-amd64
./kie --config conf/kie-conf.yaml
```