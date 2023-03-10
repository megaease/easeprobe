# EaseProbe
EaseProbe is a simple, standalone, and lightweight tool that can do health/status checking, written in Go.

## Installation

### From Source
- Start Quickly
  ```shell
  helm install easeprobe ./helm/easeprobe
  ```
  > **Note**:
  > Persistence for EaseProbe is not enabled by default, you must enable it for production environment.

- Start with Persistence
  ```shell
  helm install easeprobe ./helm/easeprobe \
    --set persistence.enabled=true
  ```

- Start with Prometheus and Grafana
  ```shell
  helm dependency build ./helm/easeprobe
  helm install [RELEASE_NAME] ./helm/easeprobe \
    --set prometheus.enabled=true \
    --set grafana.enabled=true
  ```
  > **Note**:
  > Persistence for Prometheus and Grafana is not enabled by default, you must enable it for production environment.

### From Repo
- Add repository
  ```shell
  helm repo add easeprobe https://megaease.github.io/easeprobe
  ```
- Install and run
  ```shell
  helm install [RELEASE_NAME] easeprobe/easeprobe
  ```

## Uninstallation
```shell
helm uninstall [RELEASE_NAME]
```

## Parameters
| Name | Description | Value |
| ---- | ----------- | ----- |
| `config` | Configuration for EaseProbe, refer to [Manual](https://github.com/megaease/easeprobe/blob/main/docs/Manual.md) | `{}`
| `image.repository` | Image repository | `megaease/easeprobe`
| `image.tag` | Image tag | `latest`
| `image.pullPolicy` | Image pull policy | `IfNotPresent`
| `imagePullSecrets` | Image pull secrets | `[]`
| `persistence.enabled` | Whether to enable persistence | `false`
| `persistence.existingClaim` | Existing PVC name | `""`
| `persistence.storageClassName` | Storage class name | `""`
| `persistence.size` | Volume size for persistence | `1Gi`
| `prometheus` | Configuration for Prometheus, refer to [https://artifacthub.io/packages/helm/prometheus-community/prometheus](https://artifacthub.io/packages/helm/prometheus-community/prometheus) | `{}`
| `prometheus.enabled` | Whether to enable Prometheus | `false`
| `grafana` | Configuration for Grafana, refer to [https://artifacthub.io/packages/helm/grafana/grafana](https://artifacthub.io/packages/helm/grafana/grafana) | `{}`
| `grafana.enabled` | Whether to enable Grafana | `false`
