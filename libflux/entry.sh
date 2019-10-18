#!/bin/bash -x

chown -R builder:builder /home/builder/.cache
exec runuser -u builder -- "$@"
