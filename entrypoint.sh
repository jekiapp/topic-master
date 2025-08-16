#!/bin/sh
rm -rf /app/infra/test_data/*
exec /app/topic-master "$@" 