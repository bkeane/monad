```mermaid
architecture-beta
group deployment
service lambda(logos:aws-lambda)[service] in deployment
service eventbridge(logos:aws-eventbridge)[bus] in deployment
eventbridge:R -[rule match]-> L:lambda
```