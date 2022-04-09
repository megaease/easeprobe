#!/bin/bash

pushd `dirname $0` > /dev/null
SCRIPT_PATH=`pwd -P`
popd > /dev/null

for i in $(find ./ -name '*.go')
do
  if ! grep -q Copyright $i
  then
    cat ${SCRIPT_PATH}/copyright.txt $i >$i.new && mv $i.new $i
  fi
done
