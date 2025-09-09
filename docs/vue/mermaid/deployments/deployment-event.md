```mermaid
architecture-beta
group deployment
service lambda(logos:aws-lambda)[lambda] in deployment
service eventbridge(logos:aws-eventbridge)[eventbridge] in deployment
eventbridge:R -[invoke]-> L:lambda
```