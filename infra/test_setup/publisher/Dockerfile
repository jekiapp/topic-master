FROM alpine:latest

RUN apk add --no-cache curl bash

WORKDIR /app

COPY publisher/create_topics.sh .
COPY topics.txt .
RUN chmod +x create_topics.sh

CMD ["./create_topics.sh"] 