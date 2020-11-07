FROM golang:1.15.3

RUN mkdir -p /codingchallenge
WORKDIR /codingchallenge

ADD config.json ./
COPY codingchallenge ./

EXPOSE 8080

CMD ["./codingchallenge"]