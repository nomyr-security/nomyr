FROM ubuntu:24.04@sha256:68ee22004a9109b44973313b02ff7b1de250f1de0f5ecb628d73f579aaff33ac AS build

ARG TARGETARCH

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates=20260816ubuntu0.24.04.1 \
        curl \
        libfontconfig1 \
        tar \
    && rm -rf /var/lib/apt/lists/*

RUN set -eux; \
    case "$TARGETARCH" in \
        amd64) corretto_arch=x64; corretto_sha256=a47243bd888b02a0fbd26d462caf394b756d45a5cb64ea96de74d01b37cde7cc ;; \
        arm64) corretto_arch=aarch64; corretto_sha256=e0bfcea80e052db916f0a0c73ffa50ee35265874f3da73851e806948769c71d0 ;; \
        *) echo "unsupported TARGETARCH: $TARGETARCH" >&2; exit 1 ;; \
    esac; \
    curl --fail --location --retry 3 --proto '=https' --proto-redir '=https' \
        "https://corretto.aws/downloads/resources/21.0.13.0.10/amazon-corretto-21.0.13.0.10-linux-${corretto_arch}.tar.gz" \
        --output /tmp/corretto.tar.gz; \
    echo "$corretto_sha256  /tmp/corretto.tar.gz" | sha256sum --check --status; \
    mkdir -p /opt/corretto; \
    tar -xzf /tmp/corretto.tar.gz --strip-components=1 -C /opt/corretto; \
    rm /tmp/corretto.tar.gz; \
    for alias in \
        'comodoaaaca [jdk]' \
        emsignrootca-c1 \
        emsigneccrootca-c3 \
        emsigneccrootca-g3 \
        'emsigneccrootcag3 [jdk]'; do \
        if /opt/corretto/bin/keytool -list -keystore /opt/corretto/lib/security/cacerts -storepass changeit -alias "$alias" >/dev/null 2>&1; then \
            /opt/corretto/bin/keytool -delete -keystore /opt/corretto/lib/security/cacerts -storepass changeit -alias "$alias"; \
        fi; \
    done; \
    /opt/corretto/bin/jlink \
        --add-modules ALL-MODULE-PATH \
        --strip-debug \
        --no-man-pages \
        --no-header-files \
        --compress=2 \
        --output /opt/java-runtime; \
    mkdir -p /opt/java-runtime/lib/security; \
    cp -L /opt/corretto/lib/security/cacerts /opt/java-runtime/lib/security/cacerts; \
    if /opt/corretto/bin/keytool -list -keystore /opt/java-runtime/lib/security/cacerts -storepass changeit | grep -Ei 'comodoaaaca|emsign(rootca-c1|eccrootca-c3|eccrootca-g3|eccrootcag3)'; then \
        echo 'distrusted eMudhra-related root remains in Java truststore' >&2; \
        exit 1; \
    fi

RUN set -eux; \
    curl --fail --location --retry 3 --proto '=https' --proto-redir '=https' \
        https://github.com/keycloak/keycloak/releases/download/26.8.1/keycloak-26.8.1.tar.gz \
        --output /tmp/keycloak.tar.gz; \
    echo '4e2740b181faa2ea623a1c2ecf00855935311e546046e2f857615a47ed24cbe8  /tmp/keycloak.tar.gz' | sha256sum --check --status; \
    mkdir -p /opt/keycloak; \
    tar -xzf /tmp/keycloak.tar.gz --strip-components=1 -C /opt/keycloak; \
    rm /tmp/keycloak.tar.gz

ENV JAVA_HOME=/opt/java-runtime \
    PATH=/opt/java-runtime/bin:$PATH \
    KC_RUN_IN_CONTAINER=true \
    KC_HEALTH_ENABLED=true

RUN /opt/keycloak/bin/kc.sh build

FROM ubuntu:24.04@sha256:68ee22004a9109b44973313b02ff7b1de250f1de0f5ecb628d73f579aaff33ac

RUN apt-get update \
    && apt-get upgrade --yes \
    && apt-get install -y --no-install-recommends \
        bash \
        ca-certificates \
        libfontconfig1 \
    && rm -rf /var/lib/apt/lists/* \
    && ca_version="$(dpkg-query -W -f='${Version}' ca-certificates)" \
    && dpkg --compare-versions "$ca_version" ge 20260816ubuntu0.24.04.1 \
    && if ! getent group 1000 >/dev/null; then groupadd --gid 1000 keycloak; fi \
    && if ! getent passwd 1000 >/dev/null; then \
        if getent passwd keycloak >/dev/null; then \
            echo 'keycloak account exists with a non-1000 UID' >&2; exit 1; \
        fi; \
        useradd --uid 1000 --gid 1000 --home-dir /home/keycloak --create-home --shell /usr/sbin/nologin keycloak; \
    fi \
    && install -d --owner=1000 --group=1000 /home/keycloak

COPY --from=build /opt/java-runtime /opt/java
COPY --from=build /opt/keycloak /opt/keycloak

RUN install -d --owner=1000 --group=1000 /opt/keycloak/data \
    && chown -R 1000:1000 /opt/keycloak/data

ENV JAVA_HOME=/opt/java \
    PATH=/opt/java/bin:$PATH \
    HOME=/home/keycloak \
    KC_RUN_IN_CONTAINER=true \
    KC_HEALTH_ENABLED=true \
    LANG=C.UTF-8

WORKDIR /opt/keycloak
USER 1000:1000
EXPOSE 8080 9000
ENTRYPOINT ["/opt/keycloak/bin/kc.sh"]
