# SPDX-FileCopyrightText: 2026 Nomyr technologies
# SPDX-License-Identifier: AGPL-3.0-only
FROM scratch
ARG BINARY
COPY ${BINARY} /nomyr-worker
USER 10002:10002
ENTRYPOINT ["/nomyr-worker"]
