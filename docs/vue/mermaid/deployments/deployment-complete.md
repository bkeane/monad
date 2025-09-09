```mermaid
architecture-beta
group deployment
service lambda(logos:aws-lambda)[lambda] in deployment
service eventbridge(logos:aws-eventbridge)[eventbridge] in deployment
service gateway(logos:aws-api-gateway)[api gateway] in deployment
eventbridge:R -[invoke]-> L:lambda
lambda:R <-[invoke]-> L:gateway
```