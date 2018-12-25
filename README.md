# Kubernetes Initializer for LXCFS

> Initializers are an alpha feature and subject to change. Please report any Initializer specific issues on the [Kubernetes issue tracker](https://github.com/kubernetes/kubernetes/issues).

## Build (Optional, for developer only)

Build initializer

```
build-container
```

#### Deploy  

Deploy LXCFS FUSE with DeamonSet

```
kubectl apply -f lxcfs-daemonset.yaml
```

Deploy initializer

```
kubectl apply -f lxcfs-initializer.yaml
```

#### Test

```
kubectl apply -f web.yaml
```

