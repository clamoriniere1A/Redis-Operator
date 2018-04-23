#!/bin/bash

GIT_ROOT=$(git rev-parse --show-toplevel)
. ${GIT_ROOT}/tools/demo-utils.sh

desc "Start minikube with RBAC option"
run "minikube start --extra-config=apiserver.Authorization.Mode=RBAC"

desc "Create the missing rolebinding for k8s dashboard"
run "kubectl create clusterrolebinding add-on-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default"

desc "Create the cluster role binding for the helm tiller"
run "kubectl create clusterrolebinding tiller-cluster-admin  --clusterrole=cluster-admin --serviceaccount=kube-system:default"

desc "Init the helm tiller"
run "helm init --wait"

PWD=$(pwd)
cd $GIT_ROOT
eval $(minikube docker-env)
desc "First build the container"
run "make TAG=latest container"

desc "add the helm service-catalog repo"
run "helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com"

desc "check if everything is fine"
run "helm search service-catalog"

desc "install the service-catalog helm chart"
run "helm install svc-cat/catalog  --version 0.1.9 --name catalog --namespace catalog"

desc "check service-catalog components pod creation"
run "watch kubectl get pods -n catalog"

desc "[[ Service-Catalog Up-And-Running !!!]] "

cd $PWD
