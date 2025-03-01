# FindServers Project Guide

## Project Overview
FindServers is a web application that provides a fast, ad-free server browser for Counter-Strike 2 community servers. It acts as an alternative to the in-game browser with additional features and better performance.

### Key Features
- Live community server listing from Steam API
- Filtering by name, map, region, and tags
- Server sorting and connection via steam:// protocol
- Caching system for optimal performance
- Mobile-friendly responsive design

## Architecture
- **Backend**: Go application using Gin web framework
- **Frontend**: Single-page application using Alpine.js for reactivity and HTMx for dynamic content
- **Data Source**: Steam Web API for server data
- **Caching**: In-memory server cache with automatic refresh
- **Deployment**: Docker container deployed to fly.io

## Directory Structure
- `cache/`: Server caching implementation
- `models/`: Data models including Server struct
- `steam/`: Steam API client and data fetching logic
- `static/`: Frontend assets (HTML, CSS, JS)
- `scripts/`: Utility scripts for setup and deployment
- `vendor/`: Third-party frontend dependencies (managed by setup-vendor.sh)

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

## Data Flow
1. Server fetches CS2 server list from Steam API (steam/client.go)
2. Data is filtered and cached in memory (cache/cache.go)
3. Frontend requests server list via API endpoint
4. Server generates HTML for the client (main.go)
5. Frontend renders servers and handles user interactions

## API Endpoints
- `GET /api/servers`: Returns JSON list of all servers
- `GET /api/servers/list`: Returns HTML fragment with server list for HTMx

## Code Style Guidelines
- **Formatting**: Use standard Go formatting with gofmt
- **Imports**: Group standard library imports first, then third-party, then local
- **Error Handling**: Always check errors and log appropriately
- **Naming**: Use CamelCase for exported identifiers, camelCase for unexported
- **Type Definitions**: Define models in the models/ directory
- **API Endpoints**: Use RESTful conventions in Gin routes
- **HTML Generation**: Use template.HTMLEscapeString for user-provided content
- **Frontend**: Alpine.js for reactivity, HTMx for dynamic content

## Configuration
- Steam API key should be set in `.env` file as `STEAM_API_KEY`
- Server refresh interval is set to 30 seconds
- Cache invalidation occurs after 5 minutes of inactivity

## Common Development Tasks
- **Adding a new API endpoint**: Add a new route handler in main.go
- **Modifying server data**: Update the Server struct in models/server.go
- **Changing frontend styling**: Modify static/style.css
- **Adding frontend functionality**: Update Alpine.js code in static/index.html

## Performance Considerations
- Server list is cached to minimize API calls to Steam
- Frontend uses efficient DOM operations via HTMx
- Client-side filtering to reduce server load
- Service worker for offline capability
