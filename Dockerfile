FROM golang:1 as builder

RUN mkdir /build
ADD . /build/
WORKDIR /build

RUN CGO_ENABLED=0 GOOS=linux go build -a -buildvcs=false -installsuffix cgo -ldflags "-extldflags '-static'" -o main github.com/codemicro/website/site

RUN cp ./main /main
WORKDIR /run

RUN rm -rf /build

CMD ["/main"]
