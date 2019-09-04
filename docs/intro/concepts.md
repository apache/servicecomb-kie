# Concepts

### Labels
key value must belong to a identical label,
a labels is equivalent to a map, it is represent as a json object
```json
{
"app": "some_app",
"service": "payment",
"environment": "production",
"version": "1.0.1"
}
```
### key value
A key is usually related to some function in your program, let's say a new web UI should enable or not.
the labels is just like a map to tell you where this program located in. the map says, 
it is located in production environment,the version is 1.0.1
```json
{
	"value":"true",
	"labels":{
              "service": "web-console",
              "environment": "production",
              "version": "1.0.1"
             }
}
```