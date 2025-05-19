FROM curlimages/curl:latest

COPY create_topics.sh /create_topics.sh
COPY topics.txt /topics.txt
RUN chmod +x /create_topics.sh

CMD ["/create_topics.sh"] 