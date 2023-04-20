FROM golang:1.20-alpine as builder

WORKDIR /go-src

# install required packages
RUN apk add --no-cache make gcc musl-dev linux-headers git

COPY ./scanner/ .
COPY ./service/db/migration/ ./migration/

RUN ls -laR .
RUN ls -laR ./migration/

RUN go mod download
RUN make scanner

FROM alpine:latest

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /go-src/build/bin/ .
COPY --from=builder /go-src/migration/ ./migration/
COPY --from=builder /go-src/docker-entry-point.sh .
COPY --from=builder /go-src/wait-for-it.sh .

RUN apk add --no-cache --upgrade bash

CMD [ "/app/scanner.exe" ]
ENTRYPOINT [ "/app/docker-entry.point.sh"]   


