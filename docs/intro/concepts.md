# Concepts

### labels
key value must belong to an identical label,
a labels is equivalent to a map, it is represent as a json object
```json
{
"app": "some_app",
"service": "web",
"environment": "production",
"version": "1.0.1"
}
```
for each unique label map, kie will generate a label id for it and produce a db record.
### key value
A key value is usually a snippet configuration for your component, let's say a web UI widget should be enabled or not.
But usually, a component has different version and deployed in different environments.
the labels is indicates where this component located in. 
below the labels indicates it is located in production environment,the version is 1.0.1
```json
{
 "id": "05529229-efc3-49ca-a765-05759b23ab28",
 "key": "enable_a_function",
 "value": "true",
 "value_type": "text",
 "create_revision": 13,
 "update_revision": 13,
 "labels": {
     "app": "some_app",
     "service": "web",
     "environment": "production",
     "version": "1.0.1"
  }
}
```

### revision
kie holds a global revision number it starts from 1, 
each creation or update action of key value record will cause the increasing of this revision number,
key value has a immutable attribute "create_revision" when first created.
after each modify,  "update_revision" will increase.
kie leverage this revision to reduce network cost, 
when client polling for key values, it can give a revision number "?revision=200" in query parameter, 
server side compare current revision with it , if they are the same, server will only return http status 304 to client.
at each query server will return current revision to client with response header "X-Kie-Revision"
