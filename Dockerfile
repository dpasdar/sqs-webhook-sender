# build stage
FROM golang:alpine AS build-env
ADD . /src
RUN cd /src && go build -o sender

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/sender /app/
ENTRYPOINT ["./sender"]