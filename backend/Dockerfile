FROM golang:1.26.11-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY backend ./backend

ARG SERVICE=api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/service "./backend/cmd/${SERVICE}"

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/service /usr/local/bin/service
COPY api/openapi.yaml /openapi.yaml

USER nobody
ENTRYPOINT ["/usr/local/bin/service"]
