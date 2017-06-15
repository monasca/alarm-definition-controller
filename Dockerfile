FROM alpine:3.5

ADD ./kube-alarm-definitions /

ENTRYPOINT ["/kube-alarm-definitions"]
