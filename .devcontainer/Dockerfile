FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04@sha256:d94c97dd9cacf183d0a6fd12a8e87b526e9e928307674ae9c94139139c0c6eae

ARG TARGETARCH
ENV MISE_VERSION=2026.10.17

RUN grep -Eq '^URIs: https?://((azure\.)?archive|security|ports)\.ubuntu\.com/' /etc/apt/sources.list.d/ubuntu.sources \
    && apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get upgrade --yes \
    && DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends --yes curl \
    && case "$TARGETARCH" in \
         amd64) mise_asset="mise-v${MISE_VERSION}-linux-x64.tar.gz"; mise_sha256=17565e268c5542d8ce944321e17af81e0b741b65af4aa7e8e010ee77376a4aa3 ;; \
         arm64) mise_asset="mise-v${MISE_VERSION}-linux-arm64.tar.gz"; mise_sha256=cac510e51793ad45e65fe92305bdbe4309c9c60b5907badbc7eba111738ab08f ;; \
         *) echo "unsupported mise target architecture: $TARGETARCH" >&2; exit 1 ;; \
       esac \
    && curl --proto '=https' --proto-redir '=https' --tlsv1.2 \
         --fail --location --silent --show-error \
         "https://github.com/jdx/mise/releases/download/v${MISE_VERSION}/${mise_asset}" \
         --output /tmp/mise.tar.gz \
    && printf '%s  %s\n' "$mise_sha256" /tmp/mise.tar.gz | sha256sum --check --status \
    && tar -xzf /tmp/mise.tar.gz -C /tmp \
    && install -D -m 0755 /tmp/mise/bin/mise /usr/local/bin/mise \
    && mise version | grep -F "$MISE_VERSION" \
    && mkdir -p /usr/local/share/mise \
    && printf 'mise=%s\nsource=https://github.com/jdx/mise/releases/tag/v%s\nasset=%s\nsha256=%s\n' \
         "$MISE_VERSION" "$MISE_VERSION" "$mise_asset" "$mise_sha256" \
         > /usr/local/share/mise/provenance.txt \
    && rm -f /tmp/mise.tar.gz \
    && rm -rf /tmp/mise \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

LABEL org.opencontainers.image.mise.version="${MISE_VERSION}" \
      org.opencontainers.image.mise.source="https://github.com/jdx/mise/releases/tag/v${MISE_VERSION}"

USER vscode

RUN test -w "$HOME" \
    && command -v mise \
    && mise version | grep -F "$MISE_VERSION"
