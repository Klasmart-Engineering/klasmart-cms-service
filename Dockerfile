FROM golang:1.18.0-alpine3.14
RUN mkdir /app
COPY ["*.pem", "main", "/app/"]
WORKDIR /app
CMD [ "./main" ]