#!/bin/sh

export http_proxy=${1}  https_proxy=${1} all_proxy=${1}

curl -sI ${2}
