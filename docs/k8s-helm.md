# DOS Helm

## Build Images

```sh
make build
make build-test
```

## Install
```sh
kubectl create namespace dos --dry-run=client -o yaml | kubectl apply -f -
cd deploy/helm
helm install dos ./dos -n dos
```

## Check Cluster
```sh
kubectl -n dos get pods
kubectl -n dos exec -it client -- dos node list -o json
kubectl -n dos exec -it client -- dos system leader show -o json
```

## Run Tests

The test pod runs sleep infinity, so specify the test command with kubectl exec:
```sh
kubectl -n dos exec -it test -- pytest -v tests/e2e
```

Short output:
```sh
kubectl -n dos exec -it test -- pytest -q tests/e2e
```
## Upgrade
```sh
helm upgrade dos ./dos -n dos
```
## Uninstall

```sh
helm uninstall dos -n dos
kubectl delete namespace dos
```
