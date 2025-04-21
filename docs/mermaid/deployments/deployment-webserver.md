```mermaid
architecture-beta
group deployment
service lambda(logos:aws-lambda)[service] in deployment
service gateway(logos:aws-api-gateway)[api gateway] in deployment
gateway:R <-[req resp]-> L:lambda
```