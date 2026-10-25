FROM postgres:18.6-alpine3.24@sha256:898c60cb390d1085bc97bdd6a1ae968757421df394b071848b52b1c09c89d6a5

USER root

RUN set -eux; \
    apk add --no-cache --upgrade 'brotli-libs=1.2.1-r0'; \
    test "$(apk list --installed -q brotli-libs)" = 'brotli-libs-1.2.1-r0'

USER postgres
