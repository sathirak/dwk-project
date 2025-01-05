# Dockerfile for CronJob
FROM alpine:latest

RUN apk add --no-cache bash curl

WORKDIR /usr/src/app

COPY job.sh .

RUN chmod +x job.sh

CMD ["./job.sh"]
