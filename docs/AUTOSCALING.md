# Horizontal Autoscaling

## Kubernetes HPA

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: limiter-api
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: limiter-api
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Pods
      pods:
        metric:
          name: kafka_consumer_lag
        target:
          type: AverageValue
          averageValue: 1000
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: limiter-consumer
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: limiter-consumer
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Pods
      pods:
        metric:
          name: kafka_consumer_lag
        target:
          type: AverageValue
          averageValue: 500
```

## Prometheus Metrics

- `rate_limiter_requests_total` - request rate
- `kafka_consumer_lag` - Kafka consumer lag
- `redis_hit_ratio` - Redis cache hit ratio

## Scaling Rules

- Scale up when CPU > 70% or Kafka lag > 1000
- Scale down after 5 minutes of low usage
- Max scale-up: 2 pods per minute
