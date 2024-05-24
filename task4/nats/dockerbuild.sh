set -x
eval $(minikube docker-env)
docker build -t natscustom .