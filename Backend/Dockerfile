# Base Golang image for building
FROM golang:1.23.4 AS builder

# Set the working directory
WORKDIR /app

# Build chatgpt module
COPY ./ /app/buildfiles
WORKDIR /app/buildfiles/main
RUN go mod tidy
RUN go build -o seniorlabai .
RUN mv seniorlabai /app
RUN mv .env /app
# Set the working directory in the final image
WORKDIR /app
RUN ls -al
# Expose the necessary port
EXPOSE 8468

# Set the entrypoint
ENTRYPOINT ["/app/seniorlabai"]
