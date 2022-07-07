#!/bin/bash

CSV=$1
LINE=$2
KIND=$3

echo "${KIND}:"
if [ "${KIND}" == "http" ]; then
    awk -v line="${LINE}" -F, 'NR>1 && (NR-1)<=line { printf "  - name: %s\n    url: https://%s\n",$7,$3}' ${CSV}
elif [ "${KIND}" == "tcp" ]; then
    awk -v line="${LINE}" -F, 'NR>1 && (NR-1)<=line { printf "  - name: %s\n    host: %s:443\n",$7,$3}' ${CSV}
fi

cat << EOF
notify:
  discord: 
     - name: "Test"
       webhook: "https://discord.com/api/webhooks/..."
       dry: true

settings:
  name: "EaseProbeTest"
  http:
    ip: 0.0.0.0
    port: 8888
    refresh: 2m
    log:
      file: /tmp/log/easeprobe.access.log
      self_rotate: false
  probe:
    timeout: 30s
    interval: 30s
  sla:
    schedule: "none" # daily, weekly, monthly, none
    data: "data.${KIND}.${LINE}/data.yaml"
EOF

