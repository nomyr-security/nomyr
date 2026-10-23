FROM --platform=$BUILDPLATFORM golang:1.27.3-bookworm@sha256:972b4110ba0f47539a975c31e52c64765e1ff19f2d74469ae4a595c6676e80a5 AS source

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
    dpkg-query -W -f='ca-certificates version ${Version}\n'; \
    sha256sum /etc/ssl/certs/ca-certificates.crt; \
    rm -rf /var/lib/apt/lists/*

RUN git init /src \
    && git -C /src remote add origin https://github.com/openbao/openbao.git \
    && git -C /src fetch --depth=1 origin refs/tags/v2.7.1 \
    && test "$(git -C /src rev-parse FETCH_HEAD)" = a5db72cef75c24b920ade02065b18dd8eb666bac \
    && git -C /src checkout --detach FETCH_HEAD

FROM source AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

RUN cp go.mod /tmp/openbao.mod \
    && cp go.sum /tmp/openbao.sum \
    && go mod edit -modfile=/tmp/openbao.mod \
        -require=golang.org/x/crypto@v0.57.0 \
        -require=golang.org/x/mod@v0.41.0 \
        -require=golang.org/x/net@v0.60.0 \
        -require=golang.org/x/sync@v0.23.0 \
        -require=golang.org/x/sys@v0.48.0 \
        -require=golang.org/x/term@v0.46.0 \
        -require=golang.org/x/text@v0.42.0 \
        -require=golang.org/x/tools@v0.49.0 \
    && go mod tidy -modfile=/tmp/openbao.mod \
    && commit_date="$(git show --no-show-signature -s --format=%cd --date=format:%Y-%m-%dT%H:%M:%SZ HEAD)" \
    && mkdir -p /out \
    && GOOS="$TARGETOS" GOARCH="$TARGETARCH" CGO_ENABLED=0 go build -trimpath -modfile=/tmp/openbao.mod -mod=readonly -ldflags="-X github.com/openbao/openbao/v2/internal/version.GitCommit=a5db72cef75c24b920ade02065b18dd8eb666bac -X github.com/openbao/openbao/v2/internal/version.CommitDate=$commit_date -X github.com/openbao/openbao/v2/internal/version.fullVersion=2.7.1" -o /out/bao . \
    && ln -s bao /out/vault \
    && mkdir -p /runtime/openbao/logs /runtime/openbao/file /runtime/openbao/config /runtime/home/openbao /runtime/tmp \
    && printf 'openbao:x:10001:10001:OpenBao:/home/openbao:/sbin/nologin\n' > /runtime/passwd \
    && printf 'openbao:x:10001:\n' > /runtime/group \
    && chown -R 10001:10001 /runtime/openbao /runtime/home/openbao \
    && chmod 1777 /runtime/tmp

FROM scratch

COPY --from=build /out/bao /usr/bin/bao
COPY --from=build /out/vault /usr/bin/vault
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /usr/share/doc/ca-certificates/copyright /usr/share/licenses/ca-certificates/COPYRIGHT
COPY --from=build /runtime/passwd /etc/passwd
COPY --from=build /runtime/group /etc/group
COPY --from=source /src/LICENSE /licenses/MPL-2.0.txt
COPY --from=build --chown=10001:10001 /runtime/openbao /openbao
COPY --from=build --chown=10001:10001 /runtime/home/openbao /home/openbao
COPY --from=build /runtime/tmp /tmp

ENV PATH=/usr/bin HOME=/home/openbao BAO_CONFIG_DIR=/openbao/config

LABEL org.opencontainers.image.source="https://github.com/openbao/openbao" \
    org.opencontainers.image.revision="a5db72cef75c24b920ade02065b18dd8eb666bac" \
    org.opencontainers.image.version="2.7.1"

USER 10001:10001
EXPOSE 8200
ENTRYPOINT ["/usr/bin/bao"]
CMD ["server", "-dev", "-dev-no-store-token"]
