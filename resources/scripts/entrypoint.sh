#!/bin/sh

# Support the following running mode
# 1) run easeprobe without any arguments
# 2) run easeprobe with easeprobe arguments
# 3) run the command in easeprobe container
PROBE_CONFIG=${PROBE_CONFIG:-/opt/config.yaml}

echo "Using config file: ${PROBE_CONFIG}"

# docker run megaease/easeprobe
if [ "$#" -eq 0 ]; then
   exec /opt/easeprobe
# docker run megaease/easeprobe -f config.yaml
elif [ "$1" != "--" ] && [ "$(echo $1 | head -c 1)" == "-" ] ; then
  exec /opt/easeprobe "$@"
# docker run -it --rm megaease/easeprobe /bin/sh
# docker run -it --rm megaease/easeprobe -- /bin/echo hello world
else
  exec "$@"
fi