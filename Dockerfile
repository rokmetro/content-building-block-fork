FROM public.ecr.aws/docker/library/golang:1.24-alpine as builder

ENV CGO_ENABLED=1

RUN apk add --no-cache --update libwebp-dev make gcc g++ git

RUN mkdir /content-app
WORKDIR /content-app
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
RUN make

FROM public.ecr.aws/docker/library/alpine:3.21.3

#we need timezone database
RUN apk add --no-cache tzdata

RUN apk update && \
    apk upgrade -U && \
    apk add ca-certificates ffmpeg libwebp libwebp-tools libwebp-dev

COPY --from=builder /content-app/bin/content /
COPY --from=builder /content-app/driver/web/docs/gen/def.yaml /driver/web/docs/gen/def.yaml

COPY --from=builder /content-app/driver/web/authorization_model.conf /driver/web/authorization_model.conf
COPY --from=builder /content-app/driver/web/authorization_policy.csv /driver/web/authorization_policy.csv
COPY --from=builder /content-app/driver/web/authorization_bbs_permission_policy.csv /driver/web/authorization_bbs_permission_policy.csv
COPY --from=builder /content-app/driver/web/authorization_tps_permission_policy.csv /driver/web/authorization_tps_permission_policy.csv

COPY --from=builder /content-app/vendor/github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/authorization/authorization_model_scope.conf /content-app/vendor/github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/authorization/authorization_model_scope.conf
COPY --from=builder /content-app/vendor/github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/authorization/authorization_model_string.conf /content-app/vendor/github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/authorization/authorization_model_string.conf

COPY --from=builder /etc/passwd /etc/passwd

ENTRYPOINT ["/content"]
