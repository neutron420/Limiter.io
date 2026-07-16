# Redis Cluster & Kafka Partitioning

## Redis Cluster

```yaml
# deploy/docker/docker-compose.yml
services:
  redis-node-1:
    image: redis:7-alpine
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000
    ports:
      - "6379:6379"
  redis-node-2:
    image: redis:7-alpine
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000
    ports:
      - "6380:6379"
  redis-node-3:
    image: redis:7-alpine
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000
    ports:
      - "6381:6379"
```

## Kafka Partitioning

Increase partitions for higher throughput:
```bash
kafka-topics --bootstrap-server localhost:9092 --alter --topic api_logs --partitions 6
```

Consumer group should match partition count:
```yaml
KAFKA_GROUP_ID: "analytics_consumers"
```

## Performance tuning

- Redis: Use pipeline for batch operations.
- Kafka: Use `acks=1` for throughput over durability.
- Monitor consumer lag with `kafka-consumer-groups --bootstrap-server localhost:9092 --group analytics_consumers --describe`.
