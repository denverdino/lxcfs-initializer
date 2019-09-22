# Kubernetes Initializer for LXCFS

**NOTE: Initializers has been deprecated in Kubernetes 1.14**

Please use the version with Admission Webhook in https://github.com/denverdino/lxcfs-admission-webhook

## Build (Optional, for developer only)

Build initializer

```
build-container.sh
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

