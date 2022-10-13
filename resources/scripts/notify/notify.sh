#!/bin/bash

# This is the example script for the shell notification script.
# it shows how to retrive the data in environment variables and the StdIn
echo "----------------------------------------------------"
echo "<-------- STDIN Begin ------->"
while read -r line
do
  echo "$line"
done < "${1:-/dev/stdin}"
echo "<--------- STDIN End -------->"

echo "Type: ${EASEPROBE_TYPE}"
echo "----------------------------------------------------"

if [[ ${EASEPROBE_TYPE} == "Status" ]]; then
    echo "Title: ${EASEPROBE_TITLE}"
    echo "Name: ${EASEPROBE_NAME}"
    echo "Status: ${EASEPROBE_STATUS}"
    echo "RTT: ${EASEPROBE_RTT}"
    echo "TIME: ${EASEPROBE_TIME}"
    echo "TIMESTAMP: ${EASEPROBE_TIMESTAMP}"
    echo "Message: - ${EASEPROBE_MESSAGE}"
    echo "----------------------------------------------------"
fi

echo "${EASEPROBE_JSON}"
echo "----------------------------------------------------"
echo "${EASEPROBE_CSV}"
echo "----------------------------------------------------"
env
echo "----------------------------------------------------"