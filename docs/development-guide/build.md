# How to Build

Download init and run mongodb as mentioned in get started section

Build serivcecomb-kie binary

```shell script bash
cd ${project_root_path}
build/build_binary.sh
```

This will build 4 packages for different platforms in ${servicecomb_root_path}/release/kie directory.
Choose the right one, unzip and run it.
For Example:
```shell script bash
cd ${servicecomb_root_path}/release/kie
tar -xzvf apache-servicecomb-kie--${platform}-amd64.tar.gz
cd apache-servicecomb-kie--${platform}-amd64
./kie --config conf/kie-conf.yaml
```