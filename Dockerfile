FROM golang:1.27.4-alpine3.24@sha256:0a5f7fe796599ff487605987432e7619ae45beeb2fba4c2ce0b907879932aebf AS build

WORKDIR /app

COPY . .

RUN apk add --no-cache make git && \
  go mod tidy && \
  make build

FROM alpine:3.24@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

RUN addgroup -g 10001 -S pvtr && \
  adduser -u 10001 -S pvtr -G pvtr -h /home/pvtr

WORKDIR /home/pvtr

COPY --from=build --chown=pvtr:pvtr /app/pvtr /usr/local/bin/pvtr

USER pvtr:pvtr

CMD ["pvtr", "--help"]
