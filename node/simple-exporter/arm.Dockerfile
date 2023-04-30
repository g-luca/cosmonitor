FROM golang:1.21-alpine as build
LABEL authors="g-luca"

# Set the working directory inside the container.
WORKDIR /app
# Copy the Go source code and any necessary files to the container.
COPY . .
# Build and compile the Go application.
RUN CGO_ENABLED=0 go build -o se -a -ldflags '-w -extldflags "-static"'


FROM scratch

WORKDIR /app

# copy the ca-certificate.crt from the build stage
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# copy the built app
COPY --from=build /app/se .

# Expose the port on which your Go application will listen.
EXPOSE 9090

# Specify the command to run when the container starts.
CMD ["./se"]

#docker buildx build --platform linux/amd64 -t cosmonitor .