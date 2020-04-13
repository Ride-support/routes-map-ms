FROM golang:1.12.0-alpine3.9
RUN mkdir /app
ADD . /app
WORKDIR /app

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

RUN go get -d -v github.com/gorilla/mux \
	&& go get go.mongodb.org/mongo-driver/mongo \
    && go get github.com/umahmood/haversine

RUN go build -o main .

EXPOSE 9090

CMD ["/app/main"]