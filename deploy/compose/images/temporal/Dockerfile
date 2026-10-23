FROM --platform=$BUILDPLATFORM golang:1.27.3-bookworm@sha256:972b4110ba0f47539a975c31e52c64765e1ff19f2d74469ae4a595c6676e80a5 AS build

ARG TARGETOS
ARG TARGETARCH

RUN set -eux; \
    apt-get update; \
    apt-get install --yes --no-install-recommends --reinstall ca-certificates=20250419~deb12u2; \
    test "$(dpkg-query -W -f='${Version}' ca-certificates)" = '20250419~deb12u2'; \
    sed -i \
      -e 's@^mozilla/emSign_ECC_Root_CA_-_C3\.crt$@!mozilla/emSign_ECC_Root_CA_-_C3.crt@' \
      -e 's@^mozilla/emSign_Root_CA_-_C1\.crt$@!mozilla/emSign_Root_CA_-_C1.crt@' \
      /etc/ca-certificates.conf; \
    update-ca-certificates --fresh; \
    if openssl crl2pkcs7 -nocrl -certfile /etc/ssl/certs/ca-certificates.crt \
      | openssl pkcs7 -print_certs -noout \
      | grep -Eq 'CN = emSign (Root CA - C1|ECC Root CA - C3|ECC Root CA - G3)(,|$)'; then exit 1; fi; \
    dpkg-query -W -f='ca-certificates version ${Version}\n' ca-certificates; \
    sha256sum /etc/ssl/certs/ca-certificates.crt; \
    rm -rf /var/lib/apt/lists/*

WORKDIR /src

ADD --checksum=sha256:2417fbc1348cc3b3e5b01166ebe4c40938d779505d8e119b2ed77bb3d6ae965c https://codeload.github.com/temporalio/cli/tar.gz/1de87a9f26991bf4f5c0a5ff96f2cea8d7a3cbde /tmp/temporal-cli.tar.gz
COPY deploy/compose/images/temporal/dependencies.patch /tmp/temporal-dependencies.patch

RUN set -eux; \
    tar -xzf /tmp/temporal-cli.tar.gz --strip-components=1 -C /src; \
    test "$(git hash-object /src/go.mod)" = '48cd4f572ff81cf4b3b708ce809467859a099076'; \
    test "$(git hash-object /src/go.sum)" = '28f5a717fe83cf9de23d91dc28097033a780f327'; \
    git apply --check --unidiff-zero --unsafe-paths /tmp/temporal-dependencies.patch; \
    git apply --unidiff-zero --unsafe-paths /tmp/temporal-dependencies.patch; \
    export GOFLAGS=-mod=readonly; \
    mkdir -p /out/home/.config/temporalio /out/data /out/licenses/temporal; \
    chmod 0555 /out/home /out/home/.config /out/home/.config/temporalio; \
    chmod 0700 /out/data; \
    chown 10001:10001 /out/data; \
    printf 'temporal:x:10001:10001:Temporal development user:/home/temporal:/sbin/nologin\n' > /out/passwd; \
    printf 'temporal:x:10001:\n' > /out/group; \
    cp /src/LICENSE /out/licenses/temporal/LICENSE; \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOTOOLCHAIN=local \
      go build -mod=readonly -trimpath -buildvcs=false -ldflags='-s -w' \
      -o /out/temporal /src/cmd/temporal; \
    go list -mod=readonly -deps -f '{{if .Module}}{{.Module.Path}}@{{.Module.Version}}|{{.Module.Dir}}{{end}}' /src/cmd/temporal \
      | sort -u \
      | while IFS='|' read -r module_ref module_dir; do \
          case "$module_dir" in /go/pkg/mod/*) \
            find "$module_dir" -maxdepth 1 -type f \( -iname 'LICENSE*' -o -iname 'COPYING*' -o -iname 'NOTICE*' \) \
              -exec sh -c 'module_ref=$1; shift; for license do install -D -m 0644 "$license" "/out/licenses/modules/$module_ref/${license##*/}"; done' sh "$module_ref" {} +; \
            ;; \
          esac; \
        done

FROM scratch

LABEL org.opencontainers.image.source="https://github.com/temporalio/cli" \
      org.opencontainers.image.revision="1de87a9f26991bf4f5c0a5ff96f2cea8d7a3cbde" \
      org.opencontainers.image.version="1.9.1"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /usr/share/doc/ca-certificates/copyright /usr/share/licenses/ca-certificates/COPYRIGHT
COPY --from=build /usr/share/doc/tzdata/copyright /usr/share/licenses/tzdata/COPYRIGHT
COPY --from=build /out/passwd /etc/passwd
COPY --from=build /out/group /etc/group
COPY --from=build /out/licenses /usr/share/licenses
COPY --from=build /out/home /home/temporal
COPY --chown=10001:10001 --from=build /out/data /home/temporal/data
COPY --from=build /out/temporal /usr/local/bin/temporal

ENV HOME=/home/temporal

USER 10001:10001
WORKDIR /home/temporal
EXPOSE 7233 8233
ENTRYPOINT ["temporal"]
