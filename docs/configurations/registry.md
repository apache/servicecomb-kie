# Use discovery service
kie server is able to register to a discovery service, so that other client is able to discover it registry service.
by default registry feature is disabled.
```yaml
cse:
  service:
    registry:
      disabled: true
```
For example, kie is able to register to ServiceComb service center.


### Options
this feature is powered by go-chassis,
refer to https://docs.go-chassis.com/user-guides/registry.html to know how to use

### Example 
Register to ServiceComb service center
```yaml
cse:
  service:
    registry:
      address: http://127.0.0.1:30100
...
```