### Dynamic Admission Reverse Proxy 

DARP - the K8S Admission Validation reverse proxy.

#### Build 
`make build` 

#### Upstream configuration
`config.json` example:
```
"upstreams": 
     [
       {
         # Set upstream URL 
         "url": "http://127.0.0.1:8081/validate/route",
         # Set K8S resource 
         "resource": "route"
       },
       {
         "url": "http://127.0.0.1:3000/service",
         "resource": "services"
       }
     ]
```

#### Usage example 
1. Start the server `./darp server`
2. Deploy WebHook configuration `oc create -f test/manually/webhook-cfg.yaml`
3. Start demo NodeJS server `cd test/darp_client && npm start`
4. Create good service yaml `oc create -f test/manually/good-service.yaml` - service creation allowed  
5. Create bad service yaml  `oc create -f test/manually/bad-service.yaml` - service creation declined 
