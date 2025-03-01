# FindServers Project Guide

## Build & Run Commands
- `make setup`: Install frontend dependencies
- `make dev`: Build and run with hot reloading in Docker
- `make devlog`: Same as dev but logs to findservers.log
- `make build-dev`: Build Docker image
- `make up-dev`: Start container
- `make down`: Stop and clean up
- `make clean`: Full cleanup
- `make deploy`: Deploy to fly.io
- `make logs`: View deployment logs

## Code Style Guidelines
- **Formatting**: Use standard Go formatting with gofmt
- **Imports**: Group standard library imports first, then third-party, then local
- **Error Handling**: Always check errors and log appropriately
- **Naming**: Use CamelCase for exported identifiers, camelCase for unexported
- **Type Definitions**: Define models in the models/ directory
- **API Endpoints**: Use RESTful conventions in Gin routes
- **HTML Generation**: Use template.HTMLEscapeString for user-provided content
- **Frontend**: Alpine.js for interactivity, HTMx for dynamic content