FROM docker.io/golang:1.22-alpine as builder

ENV CGO_ENABLED=1

RUN apk add --no-cache --update libwebp-dev make gcc g++

RUN mkdir /content-app
WORKDIR /content-app
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
RUN make

FROM alpine:3.17.2

#we need timezone database
RUN apk add --no-cache tzdata

RUN apk update && \
    apk upgrade -U && \
    apk add ca-certificates ffmpeg libwebp libwebp-tools libwebp-dev

COPY --from=builder /content-app/bin/content /
COPY --from=builder /content-app/docs/swagger.yaml /docs/swagger.yaml

COPY --from=builder /content-app/driver/web/authorization_model.conf /driver/web/authorization_model.conf
COPY --from=builder /content-app/driver/web/authorization_policy.csv /driver/web/authorization_policy.csv

COPY --from=builder /etc/passwd /etc/passwd

ENTRYPOINT ["/content"]
