# FindServers.net

A real-time server browser for Counter-Strike 2 built with Go and deployed at [findservers.net](https://findservers.net).

## Overview

This application provides a fast, cached interface to browse active Counter-Strike 2 servers by fetching data from the Steam Web API and serving it through a lightweight Go backend with a reactive frontend.

I built this mainly because I wanted it to exist for my own use, but would be happy if others found it useful as well. I miss seeing 10k+ active servers, and can't help feel the drop has come at least in part due to the built in server browser not being that great. There is another site that lists servers, but it only shows a few at a time and doesn't have great filters.

There is a certain feel to scrolling through and manipulating a list of servers looking for somewhere fun to play that I wanted to emulate with this app.

## Flow

```
GitHub Actions (Cron) → API Refresh Endpoint → Steam Web API → Redis Cache → Client API → Frontend
```

1. **Automated Data Collection**: GitHub Actions triggers periodic server list refreshes via authenticated API calls
2. **Steam Integration**: Backend fetches live server data from Steam Web API
3. **Data Processing**: Raw server data is cleaned and normalized using custom models
4. **Caching Layer**: Processed server data is stored in Redis for fast retrieval
5. **Client API**: Frontend requests server data through RESTful JSON API
6. **Real-time Updates**: Server list stays current through automated refresh cycles

### Endpoints

- `GET /api/servers` - Returns cached server list as JSON
- `POST /api/refresh-servers` - Triggers server data refresh (authenticated)

## Servers

Servers are processed from Steam's raw format into a consistent structure containing:
- Network details (IP, port, address)
- Server metadata (name, region, player counts)
- Game information (map, security, game modes)

## Frontend app

The front end consists only of a single page application, currently built with React (was formerly alpine.js, may consider moving back to that). The goal is to have something lightning fast, so the app is served immediately when visiting the site (no landing page).

### Filters

I also wanted to have quick and powerful filters. Filtering is handled on the front end and is more or less instantaneous. You can use both positive and negative filters, which is something the built in server browser can't do.

So, both

map: `de_, surf_`

and

map: `de_, -dust2`

are valid.

The first only shows servers that contain 'de_' or 'surf_' in the map name. The second only shows servers that contain 'de_' in the name, but excludes any with 'dust2' (thus the popular map de_dust2 is excluded, but all other de_ maps are shown).

### Scrolling

With thousands of servers generally up, scrolling can become slow. To combat this, I implemented a sort of virtual scrolling where just enough items are rendered to fill the screen, and the starting index changes when scrolled. This gives up actual movement of servers when scrolling, though they do still appear to move due to the content shifting upwards.
