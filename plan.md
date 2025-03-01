# FindServers Performance Rework Plan

## Current Performance Issues

1. **Frontend Bloat**: Alpine.js adds complexity and performance overhead
2. **Rendering Bottlenecks**: Server HTML generation happens on every request
3. **Client-side Processing**: Sorting and filtering logic runs in the browser
4. **No Performance Metrics**: Lack of visibility into load times
5. **Large Initial Payload**: All servers sent at once with no pagination

## Architecture Rework

### 1. Server-side Rendering & Caching

Move all rendering logic to the backend and implement multi-level caching:

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│                 │     │                 │     │                 │
│   Steam API     │────▶│  Server Cache   │────▶│  HTML Cache     │
│                 │     │                 │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │                 │
                                               │  Client         │
                                               │                 │
                                               └─────────────────┘
```

- **Page Cache**: Cache fully rendered pages for common views (default sort, no filters)
- **Server Element Cache**: Pre-render HTML for each server and store in cache
- **Dynamic Assembly**: Assemble HTML from cached elements based on filters/sorting

### 2. Backend Sorting & Filtering

Move all data processing to the server side:

- **Query Parameters**: Handle all filtering and sorting via URL query parameters
- **Optimized Algorithms**: Implement efficient sorting and filtering in Go
- **Pagination**: Limit initial load to 50-100 servers with pagination controls

### 3. Plain Text UI Redesign

Simplify the UI with a minimalist, text-focused design:

- **Monospace Font**: Use a consistent monospace font for clean alignment
- **ASCII/Unicode Styling**: Use simple characters for UI elements
- **Minimal CSS**: Reduce stylesheet complexity dramatically
- **No JavaScript Framework**: Replace Alpine.js with minimal vanilla JS or none at all
- **Keep HTMX**: Retain HTMX for its simple approach to dynamic content

### 4. Performance Measurement

Implement proper instrumentation:

- **Server Timing Headers**: Add Server-Timing HTTP headers
- **Logging Middleware**: Track request processing times
- **Client Metrics**: Simple timing for key user interactions

## Implementation Plan

### Phase 1: Backend Rework

1. **Cache Refactoring**
   - Modify `cache/cache.go` to include HTML generation
   - Create a new `render/server.go` for HTML templates

2. **API Endpoint Updates**
   - Update `/api/servers/list` to accept filter/sort params
   - Add pagination parameters and logic

3. **Performance Metrics**
   - Add middleware for timing in `main.go`
   - Implement Server-Timing headers

### Phase 2: Frontend Simplification

1. **UI Redesign**
   - Create new plain text server list layout
   - Design ASCII/UTF-8 UI elements for headers/controls
   - Implement minimal CSS in `static/style.css`

2. **JavaScript Reduction**
   - Remove Alpine.js dependency
   - Keep minimal HTMX for dynamic content
   - Create lightweight vanilla JS for essential interactions

3. **Pagination Controls**
   - Add next/previous page controls
   - Implement page size selection

## File Entrypoints and Changes

### Backend Changes

- **`main.go`**: Update server routes, add instrumentation
- **`cache/cache.go`**: Add HTML caching capabilities
- **`models/server.go`**: Update/simplify server model
- **NEW `render/server.go`**: Create new rendering module
- **NEW `middleware/timing.go`**: Add performance timing middleware

### Frontend Changes

- **`static/index.html`**: Completely redesign with plain text aesthetic
- **`static/style.css`**: Simplify to minimal styling focused on readability
- **NEW `static/minimal.js`**: Create minimal JS helpers if needed
- **Remove `static/vendor/alpine.min.js`**: Eliminate Alpine dependency

## Example Plain Text UI Design

```
┌── FindServers.net ────────────────────────────────────────────────────────────┐
│                                                                               │
│  Servers: 342 | Page: 1/7 | Results per page: [25] [50] [100]                 │
│                                                                               │
│  Sort by: [Name] [Players▼] [Map]                   [ Refresh ] [ Connect ]   │
│                                                                               │
│  Filters: [____________________] [Hide Empty] [US,EU] [surf,bhop,1v1,retake]  │
│                                                                               │
│  ╔════╦═══╦══════╦═══════════════════════════════════╦═════╦════════╦═══════╗ │
│  ║ PW ║VAC║Region║ Server Name                       ║Bots ║Players ║ Map   ║ │
│  ╠════╬═══╬══════╬═══════════════════════════════════╬═════╬════════╬═══════╣ │
│  ║    ║ ● ║ US   ║ BrutalCS - Retakes Mirage Only    ║     ║ 14/16  ║ mirage║ │
│  ║    ║ ● ║ EU   ║ ESEA EU Retakes (128 tick)        ║     ║ 12/12  ║ dust2 ║ │
│  ║    ║ ● ║ US   ║ Surf Utopia - Beginner Maps       ║     ║ 9/24   ║ surf_x║ │
│  ║ 🔒 ║ ● ║ EU   ║ WarmupServer.com - 1v1 Arena      ║ 2   ║ 8/32   ║ aim_m ║ │
│  ║    ║ ● ║ AS   ║ RapidNetworks Surf                ║     ║ 7/32   ║ surf_y║ │
│  ║    ║ ● ║ US   ║ Retakes Chicago 128tick            ║     ║ 5/10   ║ nuke ║ │
│  ╚════╩═══╩══════╩═══════════════════════════════════╩═════╩════════╩═══════╝ │
│                                                                               │
│  ←──  1 2 3 4 5 6 7  ──→                                                      │
│                                                                               │
│  Connection Status: Ready                                                     │
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘
```

(above is just an example, doesn't have to be exactly like that)

## Performance Goals

- **Initial Page Load**: < 200ms (cached default view)
- **Filtered View Generation**: < 100ms
- **Memory Usage**: < 100MB
- **CPU Usage**: < 10% on modest hardware

## Next Steps

1. Create a benchmark of current performance as baseline
2. Implement backend caching and rendering changes
3. Test performance improvement with current UI
4. Implement plain text UI redesign
5. Final performance testing and optimization
