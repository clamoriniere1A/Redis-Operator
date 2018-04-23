#!/bin/bash

GIT_ROOT=$(git rev-parse --show-toplevel)
export GOPATH=$GIT_ROOT/../../../../

echo "Start minikube with RBAC option"
minikube start --extra-config=apiserver.Authorization.Mode=RBAC

echo "Create the missing rolebinding for k8s dashboard"
kubectl create clusterrolebinding add-on-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default

echo "Init the helm tiller"
kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller

printf "Waiting for tiller deployment to complete."
until [ $(kubectl get deployment -n kube-system tiller-deploy -ojsonpath="{.status.conditions[?(@.type=='Available')].status}") == "True" ] > /dev/null 2>&1; do sleep 1; printf "."; done
echo

eval $(minikube docker-env)
echo "Install the redis-cluster operator"

echo "First build the container"
TAG=latest
make TAG=$TAG container
# tag the same image for rolling-update test
docker tag redisoperator/redisnode:$TAG redisoperator/redisnode:4.0

echo "create RBAC for rediscluster"
#kubectl create -f $GIT_ROOT/examples/RedisCluster_RBAC.yaml

printf  "create and install the service catalog in a dedicate namespace"
helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
until helm install --wait svc-cat/catalog  --name catalog --namespace catalog; do sleep 1; printf "."; done
echo

printf "Waiting for service catalog deployment to complete."
until [ $(kubectl get deployment catalog-catalog-apiserver -n catalog -ojsonpath="{.status.conditions[?(@.type=='Available')].status}") == "True" ] > /dev/null 2>&1; do sleep 1; printf "."; done
until [ $(kubectl get deployment catalog-catalog-controller-manager -n catalog -ojsonpath="{.status.conditions[?(@.type=='Available')].status}") == "True" ] > /dev/null 2>&1; do sleep 1; printf "."; done
echo

printf  "create and install the redis operator in a dedicate namespace"
until helm install --wait -n operator --set image.tag=$TAG --set broker.activate=true chart/redis-operator; do sleep 1; printf "."; done
echo
printf "Waiting for redis-operator deployment to complete."
until [ $(kubectl get deployment operator-redis-operator -ojsonpath="{.status.conditions[?(@.type=='Available')].status}") == "True" ] > /dev/null 2>&1; do sleep 1; printf "."; done
echo

echo "[[[ Run End2end test ]]] "
cd ./test/e2e && go test -c && ./e2e.test --kubeconfig=$HOME/.kube/config --image-tag=$TAG --ginkgo.slowSpecThreshold 260

echo "[[[ Cleaning ]]]"

echo "Remove redis-operator helm chart"
helm del --purge operator
helm del --purge catalog

kubectl delete ClusterRole redis-node
kubectl delete ClusterRoleBinding redis-node-default
