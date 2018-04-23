#!/bin/bash

GIT_ROOT=$(git rev-parse --show-toplevel)
. ${GIT_ROOT}/tools/demo-utils.sh

PWD=$(pwd)
cd $GIT_ROOT

desc "Open the k8s dashboad"
run "minikube dashboard"

desc "create and install the redis operator in a dedicate namespace"
run "kubectl create ns operators"
run "helm install -n operator chart/redis-operator --set broker.activate=true --namespace operators"

desc "check that the redisclusters is now available in the api server"
run "kubectl get redisclusters --all-namespaces"

desc "check which resources are created by the broker"
run "kubectl get clusterservicebrokers"

run "kubectl get clusterservicebrokers redis-broker -o yaml"

desc "check what are instanciated by the broker in the service-catalog"
run "watch kubectl get clusterservicebrokers,clusterserviceclasses,clusterserviceplans"

desc "see what is in the clusterserviceclasse"
run "kubectl get clusterserviceclasses -o yaml"

desc "see what is in the clusterserviceplan"
run "kubectl get clusterserviceplans -o yaml"

desc "Now, lets start the demo !!!"

desc "Create a dedicated namespace for the demo"
run "kubectl create ns test-ns"

desc="run this command in another terminal: watch kubectl -n test-ns get serviceinstance,rediscluster,pod,service"

desc "instantiate a new redis-cluster thank to the service catalog"
run "cat examples/catalog/rediscluster-instance.yaml"
run "kubectl create -n test-ns -f examples/catalog/rediscluster-instance.yaml"

desc "check what was created"
run "kubectl get -n test-ns serviceinstances"

desc "and others resources"
run "kubectl get -n operators redisclusters,pods,services,secrets"

desc "wait that the redis-cluster is ready"
run "watch kubectl plugin rediscluster -n operators"

desc "now we can create the service binding"
run "cat examples/catalog/rediscluster-binding.yaml"
run "kubectl create -n test-ns -f examples/catalog/rediscluster-binding.yaml"

desc "check if the secret was created properly"
run "kubectl get -n test-ns secrets"
run "kubectl get -n test-ns secrets rediscluster-service-binding -o yaml"

desc "unbind the redis-cluster-instance"
run "kubectl -n test-ns delete servicebinding rediscluster-service-binding"
run "kubectl get -n test-ns servicebinding,secrets"

desc "remove the redis-cluster instance"
run "kubectl -n test-ns get serviceinstance,rediscluster"
run "kubectl delete -n test-ns serviceinstance redis-cluster-instance"
run "watch kubectl -n operators get rediscluster,pod,service"

cd $PWD