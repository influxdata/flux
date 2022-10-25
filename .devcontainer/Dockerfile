FROM quay.io/influxdb/flux-build

USER root
RUN mkdir -p /home/builder/.cargo /home/builder/go/pkg && \
    chown -R builder:builder /home/builder/.cargo /home/builder/go
USER builder
