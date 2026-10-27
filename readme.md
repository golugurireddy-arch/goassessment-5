# Production-Ready Order Management System

A distributed microservices application built with Go, Gin, GORM, PostgreSQL, HashiCorp Consul, and GraphQL. The system consists of an API Gateway, an Order Service, and a Product Service with built-in resilience patterns (retries, exponential backoff, circuit breaker).

---

## Architecture Overview

```
                      Client / Postman
                            │
                            ▼
              ┌───────────────────────────┐
              │  API Gateway (Port 8000)   │
              │  - Reverse Proxy          │
              │  - JWT Auth Provider      │
              └─────────────┬─────────────┘
                            │ Service Discovery (Consul :8500)
             ┌──────────────┴──────────────┐
             ▼                             ▼
┌──────────────────────────┐  HTTP Client ┌──────────────────────────┐
│  Order Service (:8080)   │ ────────────>│ Product Service (:8081)  │
│  - REST API & GraphQL    │  (Resilient) │ - Inventory & Catalog    │
│  - GORM / PostgreSQL     │              │ - GORM / PostgreSQL      │
└────────────┬─────────────┘              └────────────┬─────────────┘
             │                                         │
             ▼                                         ▼
   PostgreSQL (orders)                       PostgreSQL (products)
```

### Microservices

1. **`api-gateway` (Port 8000)**
   - Acts as the single entry point for clients.
   - Issues signed JWT tokens (`/auth/login`).
   - Resolves downstream services dynamically via HashiCorp Consul.
   - Proxies REST and GraphQL requests to appropriate services.

2. **`order-service` (Port 8080)**
   - Manages order lifecycles (`CREATED` → `CONFIRMED` / `CANCELLED`).
   - Exposes REST endpoints (`/api/v1/orders`) and GraphQL (`/graphql`, `/playground`).
   - Calls `product-service` via an HTTP client implemented with **Timeout**, **Exponential Backoff Retries**, and a **Circuit Breaker**.
   - Includes role-based access control (RBAC) and ownership verification.

3. **`product-service` (Port 8081)**
   - Manages products, pricing, and stock status.
   - Exposes CRUD endpoints with JWT authentication and admin role enforcement.

---

## Directory Structure

```
├── .env                        # Central environment configuration
├── .env.example                # Example environment template
├── .gitignore                  # VCS ignore rules
├── register-consul.ps1         # PowerShell script to register services with Consul
├── api-gateway/
│   ├── main.go                 # Gateway router & reverse proxy
│   ├── go.mod
│   └── go.sum
├── order-service/
│   ├── main.go                 # Order service entry point
│   ├── client/                 # Resilient HTTP client (circuit breaker & retries)
│   ├── config/                 # Environment config loader
│   ├── controller/             # REST controllers
│   ├── graph/                  # GraphQL schema & resolvers
│   ├── middleware/             # Auth, RBAC, logging, recovery, request ID
│   ├── model/                  # Domain models & validation
│   ├── repository/             # GORM PostgreSQL repository
│   ├── service/                # Business logic
│   └── tests/                  # Concurrency & race tests (100 concurrent requests)
└── product-service/
    ├── main.go                 # Product service entry point
    ├── config/
    ├── controller/
    ├── middleware/
    ├── model/
    ├── repository/
    └── service/
```

---

## Prerequisites

- **Go**: `1.22+` (or matching version)
- **HashiCorp Consul**: Running locally on `http://localhost:8500`
- **PostgreSQL**: Local instance or AWS RDS instance

---

## Environment Configuration

Create a `.env` file in the root directory (or use `.env.example` as a template):

```env
# Shared Security Config
JWT_SECRET=production-secret-key-change-me

# HashiCorp Consul Service Discovery
CONSUL_HTTP_ADDR=http://localhost:8500
PRODUCT_SERVICE_NAME=product-service
ORDER_SERVICE_NAME=order-service

# API Gateway Configuration
GATEWAY_PORT=8000

# Product Service Configuration
PRODUCT_PORT=8081
PRODUCT_DB_HOST=localhost
PRODUCT_DB_PORT=5432
PRODUCT_DB_USER=postgres
PRODUCT_DB_PASSWORD=postgres
PRODUCT_DB_NAME=products
PRODUCT_DB_SSLMODE=disable

# Order Service Configuration
ORDER_PORT=8080
ORDER_DB_HOST=localhost
ORDER_DB_PORT=5432
ORDER_DB_USER=postgres
ORDER_DB_PASSWORD=postgres
ORDER_DB_NAME=orders
ORDER_DB_SSLMODE=disable
```

> **Note**: For AWS RDS, replace the DB hosts with your RDS endpoint and set `SSLMODE=require` (or connect through an SSH tunnel if private).

---

## Getting Started

### 1. Start HashiCorp Consul
```bash
consul agent -dev
```

### 2. Start PostgreSQL Databases
Ensure the databases exist before running the services:
```sql
CREATE DATABASE products;
CREATE DATABASE orders;
```

### 3. Start the Microservices

Open three separate terminals:

```bash
# Terminal 1: Product Service
cd product-service
go run main.go

# Terminal 2: Order Service
cd order-service
go run main.go

# Terminal 3: API Gateway
cd api-gateway
go run main.go
```

### 4. Register Services with Consul
Run the registration script (in PowerShell):
```powershell
./register-consul.ps1
```
Or check registered services at `http://localhost:8500`.

---

## API Endpoints

All requests should be routed through the API Gateway at `http://localhost:8000`.

### Authentication

#### Generate Token
- **POST** `/auth/login`
- **Body**:
  ```json
  {
    "userId": "user-101",
    "role": "CUSTOMER"
  }
  ```
  *(Roles can be `CUSTOMER` or `ADMIN`)*
- **Response**: Returns a JWT bearer token. Include this token in the header: `Authorization: Bearer <token>`.

---

### Products (`product-service`)

| Method | Endpoint | Access | Description |
|---|---|---|---|
| `GET` | `/products` | Authenticated | List all products |
| `GET` | `/products/:id` | Authenticated | Get product by ID |
| `POST` | `/products` | `ADMIN` only | Create a new product |

**Create Product Payload**:
```json
{
  "name": "Wireless Mouse",
  "price": 29.99,
  "inStock": true
}
```

---

### Orders (`order-service`)

| Method | Endpoint | Access | Description |
|---|---|---|---|
| `POST` | `/api/v1/orders` | Authenticated | Create a new order |
| `GET` | `/api/v1/orders/:id` | Owner / `ADMIN` | Get order details |
| `GET` | `/api/v1/orders` | `ADMIN` only | List all orders |
| `PUT` | `/api/v1/orders/:id/status` | `ADMIN` only | Update order status |

**Create Order Payload**:
```json
{
  "productId": "1",
  "quantity": 2
}
```

**Update Status Payload**:
```json
{
  "status": "CONFIRMED"
}
```
*(Valid statuses: `CREATED`, `CONFIRMED`, `CANCELLED`)*

---

### GraphQL & Playground

- **Playground UI**: `http://localhost:8000/playground`
- **GraphQL Endpoint**: `POST http://localhost:8000/graphql`

**Sample GraphQL Query**:
```graphql
query {
  orders {
    id
    customerId
    productId
    quantity
    totalAmount
    status
    createdAt
  }
}
```

---

## Running Tests

Run unit tests and concurrency race detection:

```bash
cd order-service

# Run all unit tests
go test -v ./...

# Run concurrency test (100 simultaneous order requests)
go test -v -run TestConcurrentOrders_100Requests ./tests/...

# Run with race detector (requires CGO_ENABLED=1 and gcc on Linux/macOS)
go test -v -race ./...
```