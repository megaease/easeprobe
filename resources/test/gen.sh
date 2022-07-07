#!/bin/bash

FILE=data.csv
EXTRACT=config.sh

declare -a arr=("http" "tcp")

for t in "${arr[@]}"; do
    ./${EXTRACT} ${FILE}  100   $t > config.$t.100.yaml
    ./${EXTRACT} ${FILE}  1000  $t > config.$t.1k.yaml
    ./${EXTRACT} ${FILE}  2000  $t > config.$t.2k.yaml
    ./${EXTRACT} ${FILE}  5000  $t > config.$t.5k.yaml
    ./${EXTRACT} ${FILE}  10000 $t > config.$t.10k.yaml
    ./${EXTRACT} ${FILE}  20000 $t > config.$t.20k.yaml
done
