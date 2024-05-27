set -x
eval $(minikube docker-env)
kubectl delete deployments --all
kubectl delete services --all

#(cd backend ; sh dockerbuild.sh)
#(cd frontend ; sh dockerbuild.sh)
#(cd nats ; sh dockerbuild.sh)

kubectl apply -f natsapp.yaml

minikube service frontend-service --url
