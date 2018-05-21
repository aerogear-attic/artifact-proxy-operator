FROM centos:7
WORKDIR /app
ADD ./artifact-proxy-operator /app/main
CMD ["./main"]