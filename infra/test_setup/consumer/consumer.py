#!/usr/bin/env python3
import nsq
import json
import sys
import time
import signal
import os

def message_handler(topic):
    def handler(message):
        try:
            data = json.loads(message.body.decode())
            print(f"[{topic}] Received message: {data}")
            return True
        except Exception as e:
            print(f"Error processing message for topic {topic}: {e}")
            return False
    return handler

def main():
    topics = os.environ.get('TOPICS', '').strip().split()
    if not topics:
        print("Error: TOPICS environment variable not set or empty")
        sys.exit(1)

    readers = []
    for topic in topics:
        topic = topic.strip()
        if not topic:
            continue
            
        reader = nsq.Reader(
            message_handler=message_handler(topic),
            nsqlookupd_tcp_addresses=['nsqlookupd:4161'],
            topic=topic,
            channel=f"{topic}_consumer",
            lookupd_poll_interval=15
        )
        readers.append(reader)
        print(f"Starting consumer for topic: {topic}")

    def signal_handler(signum, frame):
        print("Shutting down consumers...")
        for reader in readers:
            reader.close()
        sys.exit(0)

    signal.signal(signal.SIGTERM, signal_handler)
    signal.signal(signal.SIGINT, signal_handler)

    while True:
        time.sleep(1)

if __name__ == '__main__':
    main() 