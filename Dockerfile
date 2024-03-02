# Start from golang base image
FROM golang:alpine

RUN apk add --no-cache make
RUN apk add --no-cache git
RUN apk add --no-cache jq

# Setup folders
RUN mkdir /app
WORKDIR /app

# Copy the source from the current directory to the working Directory inside the container
COPY . .
COPY .env .

# Download all the dependencies
RUN go get -d -v ./...

# Build the Go app
RUN go build ./cmd/nourybot

# Run the executable
CMD [ "./nourybot", "jq"]
