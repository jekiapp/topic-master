FROM python:3.9-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY consumer.py .
COPY ../topics.txt /topics.txt

CMD ["python", "consumer.py"] 