FROM alpine
COPY ["main", "*.pem", "./"]
CMD ["/main"]
