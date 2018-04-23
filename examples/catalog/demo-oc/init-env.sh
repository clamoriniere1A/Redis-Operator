#!/bin/zsh

GIT_ROOT=$(git rev-parse --show-toplevel)
PWD=$(pwd)
cd $GIT_ROOT

echo "create a oc cluster with service-catalog activated"
DATA_PATH=$HOME/minishift/data
PV_PATH=$HOME/minishift/pv
mkdir -p $DATA_PATH
mkdir -p $PV_PATH
oc cluster up --host-data-dir=$DATA_PATH --host-pv-dir=$PV_PATH --use-existing-config --service-catalog
echo "login to openshift as admin"
oc login -u system:admin --insecure-skip-tls-verify=true
oc adm policy --as system:admin add-cluster-role-to-user cluster-admin admin

echo "create the helm tiller"
oc new-project tiller
export TILLER_NAMESPACE=tiller
oc project tiller
oc process -f ${GIT_ROOT}/examples/catalog/demo-oc/tiller-template.yaml -p TILLER_NAMESPACE="${TILLER_NAMESPACE}" | oc create -f -
oc policy add-role-to-user edit "system:serviceaccount:${TILLER_NAMESPACE}:tiller"
oc create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount="$TILLER_NAMESPACE":tiller


echo "open openshift dashboard"
open https://127.0.0.1:8443

export RAAS_NAMESPACE=raas-team
echo "install redis-operator in $RAAS_NAMESPACE namespace"
oc new-project $RAAS_NAMESPACE
echo "create the clusterrolebinding for the redis-operator"
oc create clusterrolebinding redis-operator-oc --clusterrole=cluster-admin --serviceaccount="$RAAS_NAMESPACE":redis-operator

echo "build containers"
make TAG=latest container
echo "helm install"
helm install -n operator chart/redis-operator --set broker.activate=true --namespace $RAAS_NAMESPACE

# add application template
oc create -f ${GIT_ROOT}/examples/catalog/demo-oc/myApp.template.yaml -n myproject