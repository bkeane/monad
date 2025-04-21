```mermaid
architecture-beta
group deployment
service laptop(bitcoin-icons:console-outline)[terminal] in deployment
service lambda(logos:aws-lambda)[service] in deployment
laptop:R -[invoke]-> L:lambda
```