#!/bin/sh

# Support the following running mode
# 1) run easeprobe without any arguments
# 2) run easeprobe with easeprobe argumetns
# 3) run the command in easeprobe container

# docker run megaease/easeprobe
if [ "$#" -eq 0 ]; then
   exec /opt/easeprobe -f /opt/config.yaml
# docker run megaease/easeprobe -f config.yaml
elif [ "$1" != "--" ] && [ "$(echo $1 | head -c 1)" == "-" ] ; then
  exec /opt/easeprobe "$@"
# docker run -it --rm megaease/easeprobe /bin/sh
# docker run -it --rm megaease/easeprobe -- /bin/echo hello world
else
  exec "$@"
fi