# SPDX-FileCopyrightText: 2026 Nomyr technologies
# SPDX-License-Identifier: AGPL-3.0-only
FROM scratch
ARG BINARY
COPY ${BINARY} /nomyr-runner
USER 10004:10004
ENTRYPOINT ["/nomyr-runner"]
