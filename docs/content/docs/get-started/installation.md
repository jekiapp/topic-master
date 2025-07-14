+++
weight = 201
date = '2025-07-10T09:10:14+07:00'
draft = true
title = 'Installation'
+++

# Installation

Topic Master is distributed as a single binary. You can download the executable for your platform from the [release page](https://github.com/jekiapp/topic-master/releases).

Make the binary executable:

```sh
chmod +x topic-master<architecture>
```

To start the application, you need to provide the HTTP address of `nsq_lookupd` and specify a data folder for Topic Master. Each `nsq_lookupd` instance should have its own `data_path`.

Example command to start the app:

```sh
./topic-master -data_path=path -nsqlookupd_http_address=http://localhost:4161
```

On the first run, you will be prompted to set a root user password, which will be stored in the database file. The application will then sync all topics to the database. You can also re-sync topics later via the UI.

After initialization is complete, the server will be available at the default port: `4181`.
