# Deployment Guide

- [Deployment Guide](#deployment-guide)
  - [1. Overview](#1-overview)
  - [2. Standalone Deployment](#2-standalone-deployment)
    - [2.1 Download EaseProbe](#21-download-easeprobe)
    - [2.2 Configure EaseProbe](#22-configure-easeprobe)
  - [2.3 Run EaseProbe](#23-run-easeprobe)
  - [3. Docker Deployment](#3-docker-deployment)
  - [4. Kubernetes Deployment](#4-kubernetes-deployment)
    - [4.1 Creating the ConfigMap for EaseProbe Configuration file](#41-creating-the-configmap-for-easeprobe-configuration-file)
    - [4.2 Creating a PV/PVC for EaseProbe SLA data persistent.](#42-creating-a-pvpvc-for-easeprobe-sla-data-persistent)
    - [4.3 Deploy EaseProbe](#43-deploy-easeprobe)
    - [4.4 Create the EaseProbe Service](#44-create-the-easeprobe-service)


## 1. Overview

Deployment of EaseProbe is a simple process. EaseProbe is a Go application and can be deployed on any platform that supports Go. EaseProbe is a single binary and can be deployed as a standalone application or in a Kubernetes pod.

## 2. Standalone Deployment

The following steps describe how to deploy EaseProbe as a standalone application.

### 2.1 Download EaseProbe

Download the latest version of EaseProbe from the [EaseProbe Release Page](https://github.com/megaease/easeprobe/releases), find the EaseProbe binary for your platform and download it.

for example,

```bash
wget https://github.com/megaease/easeprobe/releases/download/v2.0.0/easeprobe-v2.0.0-linux-amd64.tar.gz
tar -xvf easeprobe-v2.0.0-linux-amd64.tar.gz
```

Then, you will find a binary named `easeprobe` in the current `./bin` directory. You can create a symbolic link to the binary.

```bash
sudo ln -sf ${PWD}/bin/easeprobe /usr/local/bin/easeprobe
```

### 2.2 Configure EaseProbe

Before running EaseProbe, you need to create a configuration file for EaseProbe. The configuration file is a YAML file.

The following is an example of a configuration file - `/etc/easeprobe.conf`.

```yaml
http:
  - name: Google
    url: https://www.google.com

notify:
  slack:
    - name: "Slack"
      webhook: "https://hooks.slack.com/services/XXXX/BBBB/...."

settings:
  sla:
    schedule: daily
    time: 10:00:01+08:00
    data: /var/lib/easeprobe/sla #<---- data file location
  log:
    level: debug
    file: /var/log/easeprobe.log  #<---- log file location
  http:
    port: 8181
    log:
       file: /var/log/easeprobe-http-access.log #<---- access log file location

```

There are three parameters need your attention:

- `settings.sla.data`: the data file of EaseProbe
- `settings.log.file`: the log file of EaseProbe
- `settings.http.log.file`: the HTTP access log file  of EaseProbe


## 2.3 Run EaseProbe

On Linux Platform, you can configure systemd to run EaseProbe as a service. The following is an example of a systemd service file.

```ini
[Unit]
Description=EaseProbe Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/easeprobe -f /etc/easeprobe.conf
ExecStop=/bin/pkill -f easeprobe
Restart=always

[Install]
WantedBy=multi-user.target
```
Run the following command to install the service file.

```bash
sudo cp ./easeprobe.service /etc/systemd/system/easeprobe.service
sudo systemctl daemon-reload
sudo systemctl enable easeprobe.service
sudo systemctl start easeprobe.service
```

Uninstall the service file.

```bash
sudo systemctl stop easeprobe.service
sudo systemctl disable easeprobe.service
sudo rm /etc/systemd/system/easeprobe.service
```

## 3. Docker Deployment

Prepare a configuration file and a data file - `config.yaml`

```yaml
http:
  - name: Google
    url: https://www.google.com

notify:
  slack:
    - name: "Slack"
      webhook: "https://hooks.slack.com/services/XXXX/BBBB/...."

settings:
  sla:
    schedule: daily
    time: 10:00:01+08:00
```

You can run the EaseProbe by the following command.

```bash
docker run -d --name easeprobe \
  -p 8181:8181 \
  -v /path/to/config.yaml:/opt/config.yaml \
  -v /path/to/data:/opt/data \
  megaease/easeprobe:latest
```

Note:
 - `-p` option is used to expose the HTTP port of EaseProbe.
 - `-v` option is used to mount the configuration file and data file to the container.
   -  `/opt/config.yaml` is the configuration file default path in the container.
   -  `/opt/data/` is the data file default directory in the container.




## 4. Kubernetes Deployment

Because EaseProbe needs to persist data,  we have to deploy the EaseProbe as Stateful-Set in Kubernetes, this would lead a bit complex deployment process.

1. Creating the ConfigMap for EaseProbe `config.yaml`
2. Creating a PV/PVC for EaseProbe SLA data persistent.
3. Deploy EaseProbe
4. Create the EaseProbe Service

### 4.1 Creating the ConfigMap for EaseProbe Configuration file

The following is an example of a configuration file - `config.yaml`.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: easeprobe-conf
data:
  config.yaml: |
    http:
      - name: Google
        url: https://www.google.com
        failure: 2

      - name: Prometheus (in K8s)
        url: http://prometheus:9090/graph

      - name: ElasticSearch-01 (outside K8s)
        url: http://172.20.2.201:9200
        headers:
          Authorization: "Basic ABCDEFG1234asdf=="
    host:
      servers:
        - name: server1
          host: ubuntu@172.20.1.116
          key: /opt/login.pem

        - name: server2
          host: ubuntu@172.20.2.117
          key: /opt/login.pem
    notify:
      slack:
        - name: "MegaEase Slack#alert"
          webhook: "https://hooks.slack.com/services/ASDFA/BBBASD/....."
      discord:
        - name: "MegaEase Discord#alert"
          webhook: "https://discord.com/api/webhooks/212121212/....."
    settings:
      probe:
        interval: 1m
      log:
        level: "info"
```

### 4.2 Creating a PV/PVC for EaseProbe SLA data persistent.

To be simple, we use the NFS as an example

**Step One: Create NFS PV(Persistent Volume)**

> Note: You need to change the `server` and `path` to your NFS server and path.

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: easeprobe-storage-nfs
spec:
  capacity:
    storage: 10Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Recycle
  storageClassName: easeprobe-storage
  mountOptions:
    - hard
    - nfsvers=4.1
  nfs:
    path: /volumes/nfs/easeprobe
    server: 172.20.2.114
```

**Step Two: Create NFS PVC(Persistent Volume Claim)**

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: easeprobe-pvc
spec:
  accessModes:
  - ReadWriteMany
  resources:
    requests:
      storage: 1Gi
  storageClassName: easeprobe-storage
  volumeMode: Filesystem
  volumeName: easeprobe-storage-nfs
status:
  accessModes:
  - ReadWriteMany
  capacity:
    storage: 10Gi
```

### 4.3 Deploy EaseProbe

This is the deployment file for EaseProbe.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: easeprobe
  namespace: default
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: easeprobe
    spec:
      containers:
      - name: easeprobe
        image: megaease/easeprobe
        ports:
        - containerPort: 8181

      volumeMounts:
        - mountPath: /opt/config.yaml
          name: configmap-volume-0
          subPath: config.yaml
        - mountPath: /opt/data
          name: pvc-volume-easeprobe-pvc

     volumes:
      - configMap:
        name: configmap-volume-0
          name: easeprobe
          items:
          - key: config.yaml
            path: config.yaml
      - name: pvc-volume-easeprobe-pvc
        persistentVolumeClaim:
          claimName: easeprobe-pvc
```

> Note:
>
>  - The `configmap-volume-0` is the ConfigMap for `config.yaml`, which is mounted as volume under `/opt/config.yaml` in the container.
>  - The `pvc-volume-easeprobe-pvc` is the PVC for SLA data persistent, which is mounted as a volume under `/opt/data` in the container.

### 4.4 Create the EaseProbe Service

The service is used to expose the HTTP port of EaseProbe.

> Note:
>
> The following service is a ClusterIP service, you can change it to NodePort or LoadBalancer service.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: easeprobe
  namespace: default
spec:
  ports:
  - name: httpport
    port: 38181
    protocol: TCP
    targetPort: 8181
  type: ClusterIP
```