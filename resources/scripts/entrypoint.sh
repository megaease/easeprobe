#!/bin/sh

# docker run megaease/easeprobe
if [ "$#" -eq 0 ]; then
   exec /opt/easeprobe -f /opt/config.yaml
# docker run -it --rm megaease/easeprobe /bin/sh
elif [ "$(echo $1 | head -c 1)" != "-" ] ; then
  exec "$@"
# docker run megaease/easeprobe -f config.yaml
else
  exec /opt/easeprobe "$@"
fi