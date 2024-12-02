
# Web Crawler Service

This repository contains the code for a simple web crawler service that is divided into two main components: **Client** and **Server**. The client component is responsible for initiating requests to crawl web pages, while the server component handles the actual crawling and data extraction logic.

## Project Structure

```
web_crawler_service/
│
├── client/                 # Client component responsible for initiating crawling
│   ├── Dockerfile          # Dockerfile for building the client image
│   ├── go.mod              # Go module file for client dependencies
│   ├── go.sum              # Go sum file for client dependencies
│   └── main.go             # Entry point for the client component
│
├── server/                 # Server component responsible for crawling logic
│   ├── Dockerfile          # Dockerfile for building the server image
│   ├── go.mod              # Go module file for server dependencies
│   ├── go.sum              # Go sum file for server dependencies
│   └── server.go           # Entry point for the server component
│
├── deployment.yaml        # Kubernetes deployment YAML for web crawler service
└── docker-compose.yaml    # Docker Compose YAML for local development setup
```

### Client Component

The **client** is responsible for sending requests to the server to initiate the web crawling process.

- **main.go**: The entry point of the client that will make HTTP requests to the server for crawling web pages.
- **Dockerfile**: A Dockerfile to build a Docker image for the client component.
- **go.mod & go.sum**: The Go module files for managing client dependencies.

### Server Component

The **server** handles the core crawling logic and processes the requests received from the client.

- **server.go**: The entry point for the server. It starts the web crawling process, scrapes data, and serves responses.
- **Dockerfile**: A Dockerfile to build a Docker image for the server component.
- **go.mod & go.sum**: The Go module files for managing server dependencies.

## Setup and Installation

### Prerequisites

- Docker
- Docker Compose
- Kubernetes (for deployment)
- Go 1.18+ (for development)

### 1. Running with Docker Compose (Local Setup)

To run the web crawler service locally using Docker Compose, follow these steps:

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/web_crawler_service.git
   cd web_crawler_service
   ```

2. Build and start the services using Docker Compose:

   ```bash
   docker-compose up --build
   ```

   This will start the client and server components locally. You can access the client via the appropriate port and interact with the server for web crawling.


### 2. Kubernetes Deployment

To deploy the service on Kubernetes, use the provided `deployment.yaml` file:

1. Apply the Kubernetes deployment YAML:

   ```bash
   kubectl apply -f deployment.yaml
   ```

2. Verify the deployment:

   ```bash
   kubectl get pods
   kubectl get services
   ```

### 3. Docker Compose for Development

Use Docker Compose for local development and testing:

1. Start the services:

   ```bash
   docker-compose up --build
   ```

   This will spin up the client and server in separate containers with networking configured.

