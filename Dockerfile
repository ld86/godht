FROM golang
LABEL maintainer="Bogdan Melnik teh.ld86@gmail.com"

ADD . /go/src/github.com/ld86/godht
RUN go get github.com/ld86/godht/cmd/godht
RUN go install github.com/ld86/godht/cmd/godht

ENTRYPOINT ["/go/bin/godht"]
