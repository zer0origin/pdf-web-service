# 📄 pdf-web-service
**pdf-web-service** is a lightweight web server that hosts the user interface for interacting with internal PDF services. It is designed as a Single Page Application (SPA) and provides a responsive and modern user experience using **Go templates** and **HTMX**.

## Features
- Built with the [Gin](https://github.com/gin-gonic/gin) web framework
- Renders a dynamic SPA using Go's `html/template` and [HTMX](https://htmx.org/)
- Real-time toast notifications pushed to connected users.
- Connects seamlessly to internal services:
  - `pdf-api-service` – Handles PDF viewing, and database activities.
  - `pdf-data-service` – Manages PDF-related data storage and retrieval

## Architecture
The `pdf-web-service` acts as the frontend layer of the PDF platform. It delegates processing and data operations to internal services, ensuring a clear separation between presentation, business logic, and data management.

```
Client (Browser)
│
▼
pdf-web-service (UI Layer, Toast Notification, Web API)
│
├──▶ pdf-api-service (Backend logic)
│    ├──▶ Keycloak (User Management)
│    └──▶ Postgres (Data Storage)
└──▶ pdf-data-service (Data layer)
```

## Tech Stack
- **Go** – Backend language.
- **Gin** – Web framework for routing and middleware.
- **HTMX** – Enables partial updates for a dynamic user experience.
- **Go Templates** – Server-rendered HTML templates.
- **Postgres** - Database Server.
- **Keycloak** - Open source oAuth server.

## Toast Notification Endpoint
The server provides a dedicated endpoint that enables toast-style notifications. These are managed across **all active sessions 
and users**. This allows for system-wide alerts, updates, and feedback messages without requiring full page reloads.

## Development
```bash
go run main.go
````

The service will be available at: `http://localhost:8080` by default.

if you need to kill a process already running use ``lsof -i tcp:8080 | grep -o -E '[0-9]{4,}' | xargs kill -9``

> ⚠️ Note: This service depends on internal services (`pdf-api-service`, `pdf-data-service`) and will require them to be running for full functionality.

## Functionality
## User Management

### Upload documents
YOu can upload documents via the web application, once logged in.