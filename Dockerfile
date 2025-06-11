ARG GO_VERSION=1.24.2
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

ARG TARGETARCH
ARG SERVICE_NAME="app"

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/service ./cmd/${SERVICE_NAME}


FROM alpine:latest AS final

ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser

COPY --from=build /bin/service /bin/

ENTRYPOINT ["sh", "-c", "sleep 10 && /bin/service"]
