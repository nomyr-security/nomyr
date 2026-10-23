# SPDX-FileCopyrightText: 2026 Nomyr technologies
# SPDX-License-Identifier: AGPL-3.0-only
FROM scratch
ARG BINARY
COPY ${BINARY} /nomyr
USER 10001:10001
ENTRYPOINT ["/nomyr"]
