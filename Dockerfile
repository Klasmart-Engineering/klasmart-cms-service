FROM 503677224658.dkr.ecr.cn-north-1.amazonaws.com.cn/base-cms:latest
COPY main /app/
WORKDIR /app
CMD ["./main"]
