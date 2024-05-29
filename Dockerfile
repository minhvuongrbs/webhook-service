FROM golang:1.21 as service-builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY . ./
COPY go.mod .
COPY go.sum .
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o webhook_service .

ENV CONFIG_PATH ./config/local.yaml

EXPOSE 8080
EXPOSE 9090
ENTRYPOINT /app/webhook_service -config=$CONFIG_PATH
