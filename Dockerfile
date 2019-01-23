FROM golang:latest AS BUILDER

RUN mkdir -p /go/src/app
WORKDIR /go/src/app
ADD . .
RUN CGO_ENABLED=0 GO111MODULE=on go build -o clipper-ci

FROM docker:stable

COPY --from=builder /go/src/app/clipper-ci /bin/clipper-ci

EXPOSE 8080
ENTRYPOINT ["/bin/clipper-ci"]
CMD ["start"]
