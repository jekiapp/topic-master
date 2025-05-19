#!/usr/bin/env python3
import nsq
import json
import sys
import time
import signal
import os

def message_handler(message):
    try:
        data = json.loads(message.body.decode())
        print(f"[{data['event']}] Received message: {data}")
        return True
    except Exception as e:
        print(f"Error processing message: {e}")
        return False

def main():
    topic = os.environ.get('TOPIC')
    if not topic:
        print("Error: TOPIC environment variable not set")
        sys.exit(1)

    reader = nsq.Reader(
        message_handler=message_handler,
        nsqd_tcp_addresses=['nsqd:4150'],
        topic=topic,
        channel=f"{topic}_consumer",
        lookupd_poll_interval=15
    )

    def signal_handler(signum, frame):
        print(f"Shutting down consumer for topic: {topic}")
        reader.close()
        sys.exit(0)

    signal.signal(signal.SIGTERM, signal_handler)
    signal.signal(signal.SIGINT, signal_handler)

    print(f"Starting consumer for topic: {topic}")
    while True:
        time.sleep(1)

if __name__ == '__main__':
    main() 