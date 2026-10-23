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

The Corretto JDK is reduced with `jlink` while retaining all modules; it is not a pruned module selection. The build runs Keycloak's `kc.sh build` as root with `KC_HEALTH_ENABLED=true`, matching Compose's runtime health setting. Compose supplies `start-dev --http-host=0.0.0.0`, bootstrap credentials, loopback port mapping, and the `/opt/keycloak/data` named volume. The image reuses existing Ubuntu user/group IDs 1000 when present, creates them otherwise, and ensures `/home/keycloak` and the writable data directory belong to UID/GID 1000. Keycloak 26.8.1 bundles Quarkus 3.40.1; its reaugmentation build clears and rewrites `/opt/keycloak/lib/quarkus`, including `transformed-bytecode.jar`, `generated-bytecode.jar`, `quarkus-application.dat`, and `build-system.properties`. Run 39953126459 confirmed `start-dev` reaches this path and failed with `AccessDeniedException` on the transformed JAR. Only the generated `/opt/keycloak/lib/quarkus` subtree is writable by UID/GID 1000; runtime code and `/opt/keycloak/lib/lib/deployment` (including `appmodel.dat` and `deployment-class-path.dat`) remain root-owned read-only inputs. This is a dev-only mutable build-output boundary, not a production image claim. The image provides Bash for the Compose healthcheck's `/dev/tcp` probe and exposes the application and management ports, 8080 and 9000.

The output paths are verified against Quarkus 3.40.1's [`AbstractFastJarBuilder`](https://github.com/quarkusio/quarkus/blob/3.40.1/core/deployment/src/main/java/io/quarkus/deployment/pkg/jar/AbstractFastJarBuilder.java) and [`IoUtils`](https://github.com/quarkusio/quarkus/blob/3.40.1/independent-projects/bootstrap/app-model/src/main/java/io/quarkus/bootstrap/util/IoUtils.java): the reaugmentation branch empties the Quarkus output directory and writes generated archives and application metadata there. Its app-model/deployment-classpath serialization occurs only on the initial non-rebuild branch, so those files under `lib/lib/deployment` remain inputs during startup reaugmentation.

This is a local development image only. It does not enable real identity-provider authentication, credential custody, or execution capabilities.

The image has not been built or run locally because the Docker daemon is unavailable in the development environment. Hosted integration must build both architectures, scan the resulting images, and verify `java -version`, non-root execution, write access to `/opt/keycloak/data` and `/opt/keycloak/lib/quarkus`, read-only runtime code and `/opt/keycloak/lib/lib/deployment`, Keycloak startup, and the Compose healthcheck.
