FROM golang:1.20-alpine
WORKDIR /code
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o ./out/dist
CMD ./out/dist

