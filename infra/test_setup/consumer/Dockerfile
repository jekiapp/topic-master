FROM python:3.9-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY consumer.py .
COPY ../topics.txt /topics.txt

# Read topics from file and set as environment variable
RUN echo "TOPICS=$(cat /topics.txt | tr '\n' ' ')" > /app/.env

CMD ["sh", "-c", "source /app/.env && python consumer.py"] 