# Keycloak development image

This image runs the official Keycloak 26.8.1 server distribution with Amazon Corretto 21.0.13.0.10. It supports Docker `amd64` and `arm64`; the Java archive is selected and SHA-256 verified for the build target architecture.

## Pinned inputs

| Input | Pin | Provenance |
| --- | --- | --- |
| Ubuntu build and runtime base | `ubuntu:24.04@sha256:68ee22004a9109b44973313b02ff7b1de250f1de0f5ecb628d73f579aaff33ac` | Docker Hub multi-platform manifest |
| Keycloak server distribution | `26.8.1`, SHA-256 `4e2740b181faa2ea623a1c2ecf00855935311e546046e2f857615a47ed24cbe8` | [Keycloak 26.8.1 release](https://github.com/keycloak/keycloak/releases/tag/26.8.1) |
| Corretto x86-64 JDK archive | `21.0.13.0.10`, SHA-256 `a47243bd888b02a0fbd26d462caf394b756d45a5cb64ea96de74d01b37cde7cc` | [Corretto 21 release](https://github.com/corretto/corretto-21/releases/tag/21.0.13.0.10) |
| Corretto ARM64 JDK archive | `21.0.13.0.10`, SHA-256 `e0bfcea80e052db916f0a0c73ffa50ee35265874f3da73851e806948769c71d0` | [Corretto 21 release](https://github.com/corretto/corretto-21/releases/tag/21.0.13.0.10) |
| Ubuntu CA bundle | `ca-certificates >= 20260816ubuntu0.24.04.1`, checked during build after upgrading the base packages | [Ubuntu USN-8944-1](https://ubuntu.com/security/notices/USN-8944-1) |

Amazon's Corretto 21.0.13.0.10 changelog identifies OpenJDK baseline 21.0.13+9, lists Linux x86-64 and ARM64 targets, and lists CVE-2026-91415 among the fixed CVEs. The Java 21 truststore in that release still contains several roots removed after the eMudhra intermediate-key compromise. The build removes the verified `comodoaaaca [jdk]`, `emsignrootca-c1`, `emsigneccrootca-c3`, `emsigneccrootca-g3`, and `emsigneccrootcag3 [jdk]` aliases and fails if any remain. The OS bundle is also pinned to Ubuntu's fixed release.

The Corretto JDK is reduced with `jlink` while retaining all modules; it is not a pruned module selection. The build runs Keycloak's `kc.sh build` as root with `KC_HEALTH_ENABLED=true`, matching Compose's runtime health setting. Compose supplies `start-dev --http-host=0.0.0.0`, bootstrap credentials, loopback port mapping, and the `/opt/keycloak/data` named volume. The image reuses existing Ubuntu user/group IDs 1000 when present, creates them otherwise, and ensures `/home/keycloak` and the writable data directory belong to UID/GID 1000; the extracted and built installation remains root-owned. Hosted validation must confirm `start-dev` uses the prebuilt output without attempting re-augmentation writes to the installation. The image provides Bash for the Compose healthcheck's `/dev/tcp` probe and exposes the application and management ports, 8080 and 9000.

This is a local development image only. It does not enable real identity-provider authentication, credential custody, or execution capabilities.

The image has not been built or run locally because the Docker daemon is unavailable in the development environment. Hosted integration must build both architectures, scan the resulting images, and verify `java -version`, non-root execution, write access to `/opt/keycloak/data`, no attempted re-augmentation of the read-only installation, Keycloak startup, and the Compose healthcheck.
