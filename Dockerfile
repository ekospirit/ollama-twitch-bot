# Start from golang base image
FROM golang:alpine

# Setup folders
RUN mkdir /app
WORKDIR /app

# Copy the source from the current directory to the working Directory inside the container
COPY . .
COPY .env .

# Download all the dependencies
RUN go get -d -v ./...

# Build the Go app
RUN go build -o Nourybot.out .

# Run the executable
CMD [ "./Nourybot.out"]
