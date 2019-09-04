# Registry
kie server is able to register to a registry, so that other client is able to discover it registry service.
by default registry feature is disabled.
```yaml
cse:
  service:
    registry:
      disabled: true
...
```
For example, kie is able to register to ServiceComb service center.


### Options
the function is powered by go-chassis in chassis.yaml file 
refer to https://docs.go-chassis.com/user-guides/registry.html

### Example 
Register to ServiceComb service center
```yaml
cse:
  service:
    registry:
      address: http://127.0.0.1:30100
...
```