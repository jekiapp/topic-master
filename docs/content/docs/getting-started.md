---
weight: 200
date: '2025-07-10T09:09:20+07:00'
title: 'Getting Started'
sidebar:
  open: true
---
## Download

Topic Master is distributed as a single binary for various platforms. You can download the appropriate executable for your operating system and architecture from the [release page](https://github.com/jekiapp/topic-master/releases).

After downloading, make the binary executable. Replace `<architecture>` with your system's architecture (for example, `linux-amd64`, `darwin-arm64`, etc.):

```sh
chmod +x topic-master<architecture>
```

For example, on a 64-bit Linux system:

```sh
chmod +x topic-master-linux-amd64
```

## Start the Server

To start the application, you need to provide the HTTP address of your `nsq_lookupd` instance and specify a data folder for Topic Master. Each `nsq_lookupd` instance should have its own `data_path`.

Example command to start the app:

```sh
./topic-master -data_path=path/to/data -nsqlookupd_http_address=http://localhost:4161
```

On the first run, you will be prompted to set a root user password. This password will be stored in the database file. The application will then sync all topics to the database. You can also re-sync topics later via the UI.

After initialization is complete, the server will be available at the default port: `4181`.