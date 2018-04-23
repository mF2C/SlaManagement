###
FROM golang:alpine as builder

RUN apk add --no-cache git

WORKDIR /go/src/SLALite

COPY . .
RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o SLALite .

###
FROM alpine:3.6
WORKDIR /opt/slalite
COPY --from=builder /go/src/SLALite/SLALite .
COPY docker/slalite_cimi.yml /etc/slalite/slalite.yml
COPY docker/run_slalite_cimi.sh run_slalite.sh

RUN addgroup -S slalite && adduser -D -G slalite slalite
RUN chown -R slalite:slalite /etc/slalite && chmod 700 /etc/slalite

EXPOSE 8090
#ENTRYPOINT ["./run_slalite.sh"]
USER slalite
ENTRYPOINT ["/opt/slalite/SLALite"]

