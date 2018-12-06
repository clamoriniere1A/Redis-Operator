#!/bin/bash

GIT_ROOT=$(git rev-parse --show-toplevel)
export GOPATH=$GIT_ROOT/../../../../

oc cluster up --version v3.9 #--service-catalog
echo "login to openshift as admin"
oc login -u system:admin --insecure-skip-tls-verify=true
oc adm policy --as system:admin add-cluster-role-to-user cluster-admin admin

echo "create the helm tiller"
oc -n kube-system create sa tiller
oc create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --wait --service-account tiller


# Install Service Catalog
oc new-project catalog
helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
until helm install --wait svc-cat/catalog --version 0.1.9 --name catalog --namespace catalog; do sleep 1; printf "."; done; echo
until [ $($ctl get deployment catalog-catalog-apiserver -n catalog -ojsonpath="{.status.conditions[?(@.type=='Available')].status}") == "True" ] > /dev/null 2>&1; do sleep 1; printf "."; done
until [ $($ctl get deployment catalog-catalog-controller-manager -n catalog -ojsonpath="{.status.conditions[?(@.type=='Available')].status}") == "True" ] > /dev/null 2>&1; do sleep 1; printf "."; done; echo


echo "Install the redis-cluster operator"
oc project default
echo "First build the container"
make TAG=latest container
# tag the same image for rolling-update test
docker tag redisoperator/redisnode:latest redisoperator/redisnode:4.0

echo "create RBAC for rediscluster"
#oc create -f $GIT_ROOT/examples/RedisCluster_RBAC.yaml

printf  "create and install the redis operator in a dedicate namespace"
helm install --wait --namespace default -n operator --set broker.activate=true --set image.tag=latest chart/redis-operator
echo

printf "Waiting for redis-operator deployment to complete."
until [ $(oc get deployment operator-redis-operator -ojsonpath="{.status.conditions[?(@.type==\"Available\")].status}") == "True" ] > /dev/null 2>&1; do sleep 1; printf "."; done
echo

echo "[[[ Run End2end test ]]] "
cd ./test/e2e && go test -c && ./e2e.test --kubeconfig=$HOME/.kube/config --ginkgo.slowSpecThreshold 260

echo "[[[ Cleaning ]]]"

echo "wait the garbage collection"
sleep 20 
echo "Remove redis-operator helm chart"
helm del --purge operator

oc cluster down