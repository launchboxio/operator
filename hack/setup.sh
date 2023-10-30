#!/usr/bin/env bash

# Ensure minikube is up and running

# Install cluster-api and vcluster provider
clusterctl init --infrastructure vcluster

# Install crossplane and base provider packages
helm upgrade --install crossplane \
  --namespace crossplane-system \
  --create-namespace \
  crossplane-stable/crossplane

kubectl crossplane install provider crossplanecontrib/provider-kubernetes:main
kubectl crossplane install provider crossplanecontrib/provider-helm:master
