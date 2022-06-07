#!/bin/bash

echo "Type: ${TYPE}"
echo "----------------------------------------------------"

if [[ ${TYPE} == "Status" ]]; then
    echo "Title: ${TITLE}"
    echo "Name: ${NAME}"
    echo "Status: ${STATUS}"
    echo "RTT: ${RTT}"
    echo "TIMESTAMP: ${TIMESTAMP}"
    echo "Message: - ${MESSAGE}"
    echo "----------------------------------------------------"
fi

echo "${JSON}"
echo "----------------------------------------------------"
echo "${CSV}"
echo "----------------------------------------------------"