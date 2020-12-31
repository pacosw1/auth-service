FROM golang:1.15.6



RUN mkdir /web-app


ADD . /web-app

WORKDIR /web-app


RUN go build -o main .



# ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.2.1/wait /wait
# RUN chmod +x /wait




EXPOSE 9000


CMD ["/web-app/main"]
