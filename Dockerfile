FROM alpine:3.19.1 AS CA

RUN apk add --no-cache ca-certificates=20230506-r0 \
	&& addgroup -S pgweb -g 7810 && adduser -S pgweb -G pgweb -u 7810

FROM scratch

COPY --from=CA /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=CA /etc/group  /etc/group
COPY --from=CA /etc/passwd /etc/passwd

COPY ./pgweb-k8s-backend /pgweb-k8s-backend
COPY ./lib64/* /lib64/


EXPOSE 4673
USER pgweb
ENTRYPOINT ["/pgweb-k8s-backend"]
