# Etherium parser

## What does this code do?
An implementation of a simple etherium blockchain parser allowing users to subscribe to addresses and query subscriptions for transactions.

## How to use?
1. You can run the code using the vanilla Go binary with `go run cmd/blockchain_parser/main.go` and then use a tool such as Postman to hit the api
2. There is also a VSCode Launch Configuration which will allow debugging/breakpoints etc if required. See [VSCode debugging](https://code.visualstudio.com/docs/editor/debugging)
3. You can run the provided integration tests locally by using the Makefile command `make test` which will build and run the binary on `localhost:8080` and run the integration sh script. Upon completion the process is killed
4. You can run the Unit Test suite via `make test-unit`

## Further work
Due to time constraints the code is pretty basic. Further work could develop the following things:

1. API documentation using tools such as Swagger/OpenAPI
2. Address validation using regex
3. CI scripts for further automated testing such as linting or quality gateways / vulnerability testing using Sonarqube or similar
4. Container support by adding `Dockerfile` or Kubernetes manifests as required
5. More refined build steps which introspect platform architecture to build appropriate `Go` binaries
6. Increased reliability through a genuine persistance layer using tools such as Redis or a database
