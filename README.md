### sentinal-controller

Spike to see feasability of implmenting a kubernetes style controller to manage canaries.

To try this out setup minikube, then deploy the kuard app

 1. Setup minikube : https://github.com/kubernetes/minikube
 2. Deploy kuard: kubectl apply -f artifacts/examples/kuard.yaml 
```
# terminal 1
$go run *.go -kubeconfig=$HOME/.kube/config

# terminal 2
$ kubectl create -f artifacts/examples/sentinal-deployer.yml
$ kubectl apply -f artifacts/examples/sd.yml
```

Output should be as follows:

```
$kubectl get deployments
NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
kuard          2         2         2            0           11m
kuard-canary   1         1         1            0           5s
```
