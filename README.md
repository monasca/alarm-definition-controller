# Alarm Definition Controller

Build with ```CGO_ENABLED=0 go build -a -installsuffix cgo -o ./kube-alarm-definitions``` to work with included Dockerfile.

The controller is expected to run as a pod in kubernetes.

## Usage
The controller is designed to convert third party resources to alarm definitions in monasca.
It does this by reading all resources listed under ```/apis/monasca.io/v1/namespaces/<namespace>/alarmdefinitions```
 on a configurable poll loop (default 15 seconds). It is currently limited to resources created in the configured namespace, and it will only look for a monasca
endpoint in the same namespace.

See [alarm-resource.yaml](https://github.com/monasca/alarm-definition-controller/blob/master/alarm-resource.yaml) for a
yaml example of creating the resource type.

Resources must contain a valid alarm definition spec (as defined in the Monasca API documentation). If the attempt to create
the definition returns an error, it will be posted back to the original resource at the top level under "error". If the creation
request is successful, the full alarm definition will be posted to the resource. The most important part being the ID, which is
used to link the resource to the alarm definition in monasca.

See [definition-instance.yaml](https://github.com/monasca/alarm-definition-controller/blob/master/definition-instance.yaml)
for a yaml example of creating an alarm definition instance.

The controller also supports updates to the resources, but note that not all updates will be considered valid by monasca 
(see Monasca API documentation for details).
Similar to a failed creation, any errors will be posted back to the resource.

## Notes
The controller will append ``` - adc``` to the names of all alarm definitions it creates. This allows it to manage only alarm definitions
it has created and not interfere with existing definitions. It will maintain in internal cache of definitions while operating,
but requires the names to be correct on restart.

Known Issue: Errors are not cleared from the resource after successful requests.

## Examples
#### Created Valid Resource
```
ryan@ryan-HP-Z620-Workstation:~/monasca-helm/monasca$ kubectl get alarmDefinitions -o yaml
apiVersion: v1
items:
- alarmDefinitionSpec:
    expression: new_metric < 10
    name: Test Def
  apiVersion: monasca.io/v1
  kind: AlarmDefinition
  metadata:
    creationTimestamp: 2017-06-22T16:54:10Z
    name: first-alarm-definition
    namespace: default
    resourceVersion: "680748"
    selfLink: /apis/monasca.io/v1/namespaces/default/alarmdefinitions/first-alarm-definition
    uid: 698b386c-576b-11e7-9212-080027154327
kind: List
metadata: {}
resourceVersion: ""
selfLink: ""
```

Successful Response
```
ryan@ryan-HP-Z620-Workstation:~/monasca-helm/monasca$ kubectl get alarmDefinitions -o yaml
apiVersion: v1
items:
- alarmDefinitionSpec:
    deterministic: false
    expression: new_metric < 10
    id: fd6b0266-cdbc-43ac-a499-dfbfc9cd0b5a
    links:
    - href: http://monasca-api:8070/v2.0/alarm-definitions/fd6b0266-cdbc-43ac-a499-dfbfc9cd0b5a
      rel: self
    name: Test Def - adc
    severity: LOW
  apiVersion: monasca.io/v1
  kind: AlarmDefinition
  metadata:
    creationTimestamp: 2017-06-22T16:54:10Z
    name: first-alarm-definition
    namespace: default
    resourceVersion: "680756"
    selfLink: /apis/monasca.io/v1/namespaces/default/alarmdefinitions/first-alarm-definition
    uid: 698b386c-576b-11e7-9212-080027154327
kind: List
metadata: {}
resourceVersion: ""
selfLink: ""
```

#### Created Invalid Resource
```
ryan@ryan-HP-Z620-Workstation:~/monasca-helm/monasca$ kubectl get alarmDefinitions -o yaml
apiVersion: v1
items:
- alarmDefinitionSpec:
    expression: new_metric < 10
    name: Another Def
    severity: INVALID SEVERITY
  apiVersion: monasca.io/v1
  kind: AlarmDefinition
  metadata:
    creationTimestamp: 2017-06-22T16:59:12Z
    name: second-alarm-definition
    namespace: default
    resourceVersion: "681081"
    selfLink: /apis/monasca.io/v1/namespaces/default/alarmdefinitions/second-alarm-definition
    uid: 1de35dcd-576c-11e7-9212-080027154327
kind: List
metadata: {}
resourceVersion: ""
selfLink: ""
```

Failed Response
```
ryan@ryan-HP-Z620-Workstation:~/monasca-helm/monasca$ kubectl get alarmDefinitions -o yaml
apiVersion: v1
items:
- alarmDefinitionSpec:
    expression: new_metric < 10
    name: Another Def
    severity: INVALID SEVERITY
  apiVersion: monasca.io/v1
  error: 'Error: 422 {"description":"not a valid value for dictionary value @ data[u''severity'']","title":"Unprocessable
    Entity"}'
  kind: AlarmDefinition
  metadata:
    creationTimestamp: 2017-06-22T16:59:12Z
    name: second-alarm-definition
    namespace: default
    resourceVersion: "681166"
    selfLink: /apis/monasca.io/v1/namespaces/default/alarmdefinitions/second-alarm-definition
    uid: 1de35dcd-576c-11e7-9212-080027154327
kind: List
metadata: {}
resourceVersion: ""
selfLink: ""
```

## Building Type Changes
If any changes are made to the alarm definition type under pkg, follow these steps

Setup:
1) Ensure https://github.com/kubernetes/code-generator is present in the vendor directory. It may need to be manually added.
 This is required to run the code generation scripts.
2) Run ```./hack/update-codegen.sh``` from the main repo directory.
