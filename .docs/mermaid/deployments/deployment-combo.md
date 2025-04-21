```mermaid
architecture-beta
group deployment
service lambda(logos:aws-lambda)[service] in deployment
service eventbridge(logos:aws-eventbridge)[bus] in deployment
service gateway(logos:aws-api-gateway)[api gateway] in deployment
eventbridge:R -[rule match]-> L:lambda
lambda:R <-[req resp]-> L:gateway
```