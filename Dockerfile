FROM alpine
WORKDIR /app
COPY sqs-webhook-sender /sqs-webhook-sender
ENTRYPOINT ["/sqs-webhook-sender"]
