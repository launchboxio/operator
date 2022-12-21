# LaunchboxHQ Operator 

## Demo Installation 

##### Start Minikube
```bash 
minikube start --memory=8192 --cpus=4
minikube addons enable ingress
minikube tunnel
```

##### Start Operator 
```bash 
git clone git@github.com:launchboxio/operator
cd operator
make install run
```

##### Install Custom resources 
```bash 
kubectl apply -f config/samples/core_v1alpha1_cluster.yaml
kubectl apply -f config/samples/core_v1alpha1_space.yaml
kubectl apply -f config/samples/core_v1alpha1_servicecatalog.yaml
```

Once the above are completed, you should have 
 - Minikube, with metrics server installed on the host cluster 
 - A `vcluster` instance named `sample-space` 
 - Inside `vcluster`, you should have redis and mysql installed as addons
 - Lastly, the servicecatalog should be installed, and accessible on X.X.X.X 

This verifies we're able to install cluster / space addons, as well as provision services 
with addons "attached" to them