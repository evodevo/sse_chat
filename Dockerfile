FROM golang:1.14.2-alpine3.11

ARG app_env
ENV APP_ENV $app_env

RUN apk add --no-cache git

# install dependency tool
#RUN go get -u github.com/golang/dep/cmd/dep

# Fresh for rebuild on code change, no need for production
#RUN go get -u github.com/pilu/fresh

#COPY . /go/src/github.com/monstar-lab/fr-circle-api

WORKDIR /go/src/sse_chat

#CMD dep ensure && fresh

# for production, it just builds and runs the binary
# CMD go build && ./fr-circle-api

EXPOSE 3000