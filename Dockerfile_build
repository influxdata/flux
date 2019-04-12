# This file describes an image that is capable of building Flux.

FROM golang:1.12

RUN apt-get update && apt-get install -y --no-install-recommends \
		ruby \
		ragel=6.9-1.1+b1 \
	&& rm -rf /var/lib/apt/lists/*
