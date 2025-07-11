---
weight: 10
title: "Overview"
description: ""
icon: "article"
date: "2025-07-10T05:19:39+07:00"
lastmod: "2025-07-10T05:19:39+07:00"
draft: true
toc: true
---

## Problem statement
message queue is one of the crucial component in the modern software infrastructure. 
Developer on daily basis will operate with message queue provider in their development lifecycle. 
whether it is debugging, testing or sampling message.
But the built-in tools provided is less often sufficient to fulfill the software engineer operational needs. 
for example, publishing and sampling message is one of the most common operation for development software.

On the otherhand, opening developer for a full access to the MQ tools could open recurity risk.
sampling message is okay, but dumping messages to local drive is not okay. 
Deleting topics in production is also very risky if didnt protect the action.
thats why production cluster must be secured.
But the security measure built-in in the MQ platform is either very minimum or non-existent.
The security choice for infra structure manager is usually all or nothing for the developer access to the tools like NSQAdmin or Kafka-ui.

## Why Topic Master?
Based on the previous problem. Topic Master vision is to give a developer reliable tool for doing daily development job related with the MQ operation.
Topic Master will be developer centric, the feature in Topic master will only focused on developer tool (e.g publish, tail message), 
not infra tools (e.g monitoring, alerting)  because I believe there are way better tools for observability.

The security control designed to be efficient yet simple. becase we give the administration job distributed to each team. 
so the main administrator (root) will not burdened with access approval.
