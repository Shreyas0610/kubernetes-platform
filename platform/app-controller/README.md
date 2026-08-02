# App Controller

This controller adds the first Kubernetes-native platform API for the project.

Instead of asking a user to write a `Deployment`, `Service`, and `Ingress`, the platform exposes one higher-level resource:

```yaml
apiVersion: platform.sarige.dev/v1alpha1
kind: App
metadata:
  name: demo-api
spec:
  image: nginx:1.27
  port: 80
  replicas: 2
  host: demo.local
```

The controller reconciles that desired state into:

- `Deployment`
- `Service`
- optional `Ingress`
- `status.phase`
- `status.url`
- Kubernetes-style status conditions

## Local Development

```bash
make manifests
make test
make install
make run
```

In another terminal:

```bash
kubectl apply -f config/samples/platform_v1alpha1_app.yaml
kubectl get app demo-api -o yaml
kubectl get deployment,service,ingress
```

## Why This Matters

This is the Kubernetes-native layer of the platform. It demonstrates CRDs, controllers, reconciliation loops, owner references, generated manifests, RBAC, and status conditions.

The broader platform goal is to let a developer describe one high-level `App` and let the controller manage the lower-level Kubernetes resources consistently.

