FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04@sha256:d94c97dd9cacf183d0a6fd12a8e87b526e9e928307674ae9c94139139c0c6eae

RUN grep -Eq '^URIs: https?://((azure\.)?archive|security|ports)\.ubuntu\.com/' /etc/apt/sources.list.d/ubuntu.sources \
    && apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get upgrade --yes \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
