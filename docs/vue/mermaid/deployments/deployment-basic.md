```mermaid
architecture-beta
group deployment
service laptop(bitcoin-icons:console-outline)[aws cli] in deployment
service lambda(logos:aws-lambda)[lambda] in deployment
laptop:R -[invoke]-> L:lambda
```