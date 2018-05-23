FROM centos:7
WORKDIR /app
ADD ./artifact-proxy-operator /app/main
EXPOSE 8080
CMD ["./main"]