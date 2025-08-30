# FROM golang:alpine

# WORKDIR /app
# COPY . .

# RUN go mod download
# RUN go build -o main .

# EXPOSE 8080

# CMD ["./main"]

FROM golang:1.24.4-alpine


WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080
CMD ["./main"]

