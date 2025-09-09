```mermaid
architecture-beta
group deployment
service lambda(logos:aws-lambda)[lambda] in deployment
service gateway(logos:aws-api-gateway)[api gateway] in deployment
gateway:R <-[invoke]-> L:lambda
```