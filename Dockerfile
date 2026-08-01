# Single Stage build 

# 1.Base image
FROM golang:1.26
# 2.Set working directory
WORKDIR /app
# 3.Copy dependency manifests
COPY go.mod ./
# 4.copy application code to the working directory 
COPY . .
# 5.Build Go application into executable 'main' 
RUN go build -o main
# 6. Expose the port the app runs on
EXPOSE 8080
# 7. Run the executable
CMD ["/app/main"]