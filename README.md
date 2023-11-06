# LaunchboxHQ Operator 

## Installing on Workload cluster 

``` 
helm install launchboxio/operator -n lbx-system 
kubectl create secret generic \
  launchboxhq-credentials \
  --from-literal=clientId=$clientId \
  --from-literal=clientSecret=$clientSecret \
  -n lbx-system

kubectl apply -f /my/custom/cluster.yaml
```