FROM golang:1.14 as builder

WORKDIR /stale-feature-branch-operator
COPY . /stale-feature-branch-operator

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o operator pkg/*.go

FROM gcr.io/distroless/static:nonroot

WORKDIR /stale-feature-branch-operator
COPY --from=builder /stale-feature-branch-operator /stale-feature-branch-operator

USER nonroot:nonroot

ENTRYPOINT ["./operator"]
