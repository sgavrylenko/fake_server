####Fake Server
The main goal of FakeServer is provide fake app during playing with Kubernetes. 
Any request to the root location return dump of original request. 
Also it provides healthiness, readiness probes and prometheus metrics

####Run option
**--listen-address** - listen requests on. Default value ":8888"

**--ttl** - how long app will be healthy. After ttl health prob will retrun error

####Locations
"/" - return dump of any original request

"/healthz" - emulate health probe like in true app. Return ok during ttl time

"/readiness" - return ok in 5 seconds after starting app

"/metrics" - return app metrics in prometheus format

####Build and run
```shell script
docker build -t fakeserver .
```
```shell script
docker run --rm -p 8888:8888 fakeserver
```