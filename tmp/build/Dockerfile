FROM alpine:3.6

RUN adduser -D database-k8s-operator
USER database-k8s-operator

ADD tmp/_output/bin/database-k8s-operator /usr/local/bin/database-k8s-operator
