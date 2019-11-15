FROM golang:latest as builder

LABEL maintainer="Obed Tandadjaja <obed.tandadjaja@gmail.com>"

ENV APP_HOME /api-gateway
RUN mkdir $APP_HOME
WORKDIR $APP_HOME

ADD go.mod $APP_HOME/
ADD go.sum $APP_HOME/

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM scratch
COPY --from=builder /api-gateway/main .
ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/ca-bundle.pem

# Expose port 3000 to the outside world
EXPOSE 3000

# Command to run the executable
CMD ["./main"]
