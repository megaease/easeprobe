#!/bin/sh

SERVER=${1}
PORT=${2}

echo stat | nc ${SERVER} ${PORT} 