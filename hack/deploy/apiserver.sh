#!/bin/bash
set -eou pipefail

echo "checking kubeconfig context"
kubectl config current-context || { echo "Set a context (kubectl use-context <context>) out of the following:"; echo; kubectl config get-contexts; exit 1; }
echo ""

# ref: https://stackoverflow.com/a/27776822/244009
case "$(uname -s)" in
    Darwin)
        curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.1.0/onessl-darwin-amd64
        chmod +x onessl
        export ONESSL=./onessl
        ;;

    Linux)
        curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.1.0/onessl-linux-amd64
        chmod +x onessl
        export ONESSL=./onessl
        ;;

    CYGWIN*|MINGW32*|MSYS*)
        curl -fsSL -o onessl.exe https://github.com/kubepack/onessl/releases/download/0.1.0/onessl-windows-amd64.exe
        chmod +x onessl.exe
        export ONESSL=./onessl.exe
        ;;
    *)
        echo 'other OS'
        ;;
esac

# http://redsymbol.net/articles/bash-exit-traps/
function cleanup {
    rm -rf $ONESSL ca.crt ca.key server.crt server.key
}
trap cleanup EXIT

# ref: https://stackoverflow.com/a/7069755/244009
# ref: https://jonalmeida.com/posts/2013/05/26/different-ways-to-implement-flags-in-bash/
# ref: http://tldp.org/LDP/abs/html/comparison-ops.html

export KUBEPACK_NAMESPACE=kube-system
export KUBEPACK_SERVICE_ACCOUNT=default
export KUBEPACK_ENABLE_RBAC=false
export KUBEPACK_RUN_ON_MASTER=0
export KUBEPACK_DOCKER_REGISTRY=kubepack
export KUBEPACK_IMAGE_PULL_SECRET=
export KUBEPACK_UNINSTALL=0

show_help() {
    echo "packserver.sh - install kubepack apiserver"
    echo " "
    echo "packserver.sh [options]"
    echo " "
    echo "options:"
    echo "-h, --help                         show brief help"
    echo "-n, --namespace=NAMESPACE          specify namespace (default: kube-system)"
    echo "    --rbac                         create RBAC roles and bindings"
    echo "    --docker-registry              docker registry used to pull voyager images (default: kubepack)"
    echo "    --image-pull-secret            name of secret used to pull voyager operator images"
    echo "    --run-on-master                run voyager operator on master"
    echo "    --uninstall                    uninstall kubepack apiserver"
}

while test $# -gt 0; do
    case "$1" in
        -h|--help)
            show_help
            exit 0
            ;;
        -n)
            shift
            if test $# -gt 0; then
                export KUBEPACK_NAMESPACE=$1
            else
                echo "no namespace specified"
                exit 1
            fi
            shift
            ;;
        --namespace*)
            export KUBEPACK_NAMESPACE=`echo $1 | sed -e 's/^[^=]*=//g'`
            shift
            ;;
        --rbac)
            export KUBEPACK_SERVICE_ACCOUNT=packserver
            export KUBEPACK_ENABLE_RBAC=true
            shift
            ;;
        --docker-registry*)
            export KUBEPACK_DOCKER_REGISTRY=`echo $1 | sed -e 's/^[^=]*=//g'`
            shift
            ;;
        --image-pull-secret*)
            secret=`echo $1 | sed -e 's/^[^=]*=//g'`
            export KUBEPACK_IMAGE_PULL_SECRET="name: '$secret'"
            shift
            ;;
        --run-on-master)
            export KUBEPACK_RUN_ON_MASTER=1
            shift
            ;;
        --uninstall)
            export KUBEPACK_UNINSTALL=1
            shift
            ;;
        *)
            show_help
            exit 1
            ;;
    esac
done

if [ "$KUBEPACK_UNINSTALL" -eq 1 ]; then
    kubectl delete deployment -l app=kubepack --namespace $KUBEPACK_NAMESPACE
    kubectl delete service -l app=kubepack --namespace $KUBEPACK_NAMESPACE
    kubectl delete secret -l app=kubepack --namespace $KUBEPACK_NAMESPACE
    kubectl delete validatingwebhookconfiguration -l app=kubepack --namespace $KUBEPACK_NAMESPACE
    kubectl delete mutatingwebhookconfiguration -l app=kubepack --namespace $KUBEPACK_NAMESPACE
    kubectl delete apiservice -l app=kubepack --namespace $KUBEPACK_NAMESPACE
    # Delete RBAC objects, if --rbac flag was used.
    kubectl delete serviceaccount -l app=kubepack --namespace $KUBEPACK_NAMESPACE
    kubectl delete clusterrolebindings -l app=kubepack --namespace $KUBEPACK_NAMESPACE
    kubectl delete clusterrole -l app=kubepack --namespace $KUBEPACK_NAMESPACE
    kubectl delete rolebindings -l app=kubepack --namespace $KUBEPACK_NAMESPACE
    kubectl delete role -l app=kubepack --namespace $KUBEPACK_NAMESPACE

    exit 0
fi

env | sort | grep KUBEPACK*
echo ""

# create necessary TLS certificates:
# - a local CA key and cert
# - a apiserver key and cert signed by the local CA
$ONESSL create ca-cert
$ONESSL create server-cert server --domains=packserver.$KUBEPACK_NAMESPACE.svc
export SERVICE_SERVING_CERT_CA=$(cat ca.crt | $ONESSL base64)
export TLS_SERVING_CERT=$(cat server.crt | $ONESSL base64)
export TLS_SERVING_KEY=$(cat server.key | $ONESSL base64)
export KUBE_CA=$($ONESSL get kube-ca | $ONESSL base64)
rm -rf $ONESSL ca.crt ca.key server.crt server.key

curl -fsSL https://raw.githubusercontent.com/kubepack/packserver/master/hack/deploy/operator.yaml | envsubst | kubectl apply -f -

if [ "$KUBEPACK_ENABLE_RBAC" = true ]; then
    kubectl create serviceaccount $KUBEPACK_SERVICE_ACCOUNT --namespace $KUBEPACK_NAMESPACE
    kubectl label serviceaccount $KUBEPACK_SERVICE_ACCOUNT app=kubepack --namespace $KUBEPACK_NAMESPACE
    curl -fsSL https://raw.githubusercontent.com/kubepack/packserver/master/hack/deploy/rbac-list.yaml | envsubst | kubectl auth reconcile -f -
    curl -fsSL https://raw.githubusercontent.com/kubepack/packserver/master/hack/deploy/user-roles.yaml | envsubst | kubectl auth reconcile -f -
fi

if [ "$KUBEPACK_RUN_ON_MASTER" -eq 1 ]; then
    kubectl patch deploy packserver -n $KUBEPACK_NAMESPACE \
      --patch="$(curl -fsSL https://raw.githubusercontent.com/kubepack/packserver/master/hack/deploy/run-on-master.yaml)"
fi
