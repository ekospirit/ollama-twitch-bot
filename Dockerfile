# Start from a golang base image with build tools installed
FROM golang:alpine

# Install dependencies necessary for Go build
RUN apk add --no-cache --virtual .build-deps gcc musl-dev

# Setup working directory
WORKDIR /app

# Copy the source and environment file
COPY . .
COPY .env .

# Initialize Go modules and fetch dependencies
RUN go mod tidy

# Build the Go app
RUN go build -o OllamaTwitchBot.out .

# Clean up build dependencies
RUN apk del .build-deps

# Run the executable
CMD ["./OllamaTwitchBot.out"]
