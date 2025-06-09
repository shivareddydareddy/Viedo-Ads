# Viedo-Ads
This repo haing a task to track viedo ad analytics using kafka.

This project is a microservice-based system for tracking and analyzing ad interactions.It uses:

Fiber (Go) – Lightweight REST API
Kafka – Message broker for async event handling
MongoDB – single truth of data
Redis – High-performance cache and real-time analytics store
Docker + Docker Compose – For running everything locally

**flow:**
Client  --->  Fiber API |     REST API     |
                        +--------+---------+
                                 |
           +---------------------+----------------------+
           |                    |                       |
       GET /ads           POST /ads/click       GET /ads/analytics
           |                    |                       |
    Read from Redis (cache)     |                       |
    ↓ Fall back to Mongo        |                       |
                                |                       |
                           Publish to Kafka           |
                                |                       |
                          Kafka Consumer (Go worker)    |
                                |                       |
         +----------------------+------------------------+
         |                      |                        |
      Write to Mongo       Update Redis       
     (click event)       (counters: CTR etc)   

. 
