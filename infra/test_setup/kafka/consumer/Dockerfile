FROM bitnami/kafka:latest

USER root
WORKDIR /app

COPY consume_topics.sh .
COPY topics.txt ./topics.txt
RUN chmod +x consume_topics.sh

CMD ["./consume_topics.sh"] 