#!/bin/bash
while true; do
  echo "------------------"
  #netstat -na | grep TIME_WAIT  |wc
  netstat  -nat  |  awk  '{print  $6}'  |  sort  |  uniq  -c  |  sort  -n
  sleep 5
done
