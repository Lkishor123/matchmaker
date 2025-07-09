# Astrology Matchmaking Platform - Low-Level Design (LLD)

## 1. Introduction

This document provides a detailed low-level design for the services and components outlined in the High-Level Design (HLD). It specifies API endpoints, data schemas, and implementation details for each microservice, assuming the specified technology stack.

## 2. Technology Stack

* **Language:** Go
* **Web Framework:** Gin
* **ORM:** GORM
* **Relational Database:** PostgreSQL
* **Document Database:** MongoDB
* **Cache:** Redis
* **Messaging (Future):** Apache Kafka
* **Frontend:** React
* **Deployment:** Docker & Kubernetes

---

## 3. Service-Level Design

### 3.1. Auth Service

* **Responsibility:** Manages user authentication via Google OIDC and issues internal JWTs.
* **API Endpoints (Gin):**
    * `GET /api/v1/auth/google/login`: Redirects the client to Google's OAuth 2.0 consent screen.
    * `GET /api/v1/auth/google/callback`: Handles the callback from Google after user authentication.
* **Implementation Logic (`/callback`):**
    1.  Use the `golang.org/x/oauth2` library to handle the OIDC flow.
    2.  Exchange the received `authorization_code` for an access token from Google.
    3.  Call Google's user info endpoint to get the user's email and name.
    4.  Make an internal gRPC or HTTP call to the **User Service** to either retrieve or create the user (`POST /internal/v1/users`).
    5.  Generate a JWT using a library like `jwt-go`.
    6.  **JWT Claims:**
        ```json
        {
          "user_id": 123,
          "email": "user@example.com",
          "roles": ["user"],
          "exp": 1678886400, // Expiration timestamp
          "iat": 1678882800  // Issued at timestamp
        }
        ```
    7.  Sign the JWT with a securely stored RSA private key.
    8.  Return the JWT to the client.

### 3.2. User Service

* **Responsibility:** Manages core user profile data.
* **API Endpoints (Gin):**
    * `POST /internal/v1/users`: (Internal only) Creates a new user. Called by Auth Service.
    * `GET /api/v1/users/me`: Fetches the profile of the currently authenticated user (ID extracted from JWT).
    * `PUT /api/v1/users/me`: Updates the user's profile (location, interests, photo).
* **Database Schema (PostgreSQL with GORM):**
    ```go
    package models

    import (
        "gorm.io/gorm"
        "time"
    )

    // BirthDetail stores the immutable birth information for a user.
    type BirthDetail struct {
        gorm.Model
        UserID      uint      `gorm:"uniqueIndex;not null"`
        DOB         time.Time `gorm:"not null"`
        TOB         string    `gorm:"type:varchar(8);not null"` // "HH:MM:SS"
        Latitude    float64   `gorm:"type:decimal(10,8);not null"`
        Longitude   float64   `gorm:"type:decimal(11,8);not null"`
    }

    // User represents the main user profile.
    type User struct {
        gorm.Model
        Email       string      `gorm:"type:varchar(100);uniqueIndex;not null"`
        Gender      string      `gorm:"type:varchar(10)"`
        Location    string      `gorm:"type:varchar(100)"`
        PhotoURL    string
        BirthDetail BirthDetail `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
    }
    ```

### 3.3. Astrology Report Service

* **Responsibility:** A high-performance, multi-level caching wrapper for the external Astrology Engine.
* **API Endpoints (Internal Only):**
    * `POST /internal/v1/reports`: Expects birth details (`dob`, `tob`, `lat`, `lon`), returns the full JSON report.
* **Implementation Logic (Multi-Level Caching):**
    1.  Generate a stable `report_key` using `sha256(fmt.Sprintf("%s:%s:%.8f:%.8f", dob, tob, lat, lon))`.
    2.  **L1 Cache Check:** `GET report_key` from Redis. On hit, return data.
    3.  **L2 Cache Check:** On L1 miss, query MongoDB: `db.reports.findOne({_id: report_key})`. On hit, write the result to Redis (`SET report_key value EX 3600`) and return data.
    4.  **Origin Fetch:** On L2 miss, make an HTTP request to the external Astrology Engine.
    5.  **Cache Write-Back:** On successful fetch, asynchronously (in a separate goroutine) write the report to MongoDB and Redis before returning the response to the caller.
* **Schema (MongoDB):**
    ```json
    {
      "_id": "<report_key_hash>",
      "report": { /* Full JSON object from external engine */ },
      "createdAt": ISODate("...")
    }
    ```

### 3.4. Match Analysis Service

* **Responsibility:** Computes compatibility between two astrological reports.
* **API Endpoints (Gin):**
    * `POST /api/v1/analysis`: Expects a JSON body with two sets of birth details.
* **Implementation Logic:**
    1.  Define a struct for the request body: `type AnalysisRequest struct { PersonA BirthDetails; PersonB BirthDetails }`.
    2.  Use a `sync.WaitGroup` and two goroutines to concurrently call the `Astrology Report Service` for PersonA and PersonB.
    3.  Wait for both calls to complete. Handle any errors.
    4.  Pass the two retrieved JSON reports to a pure Go function `calculateCompatibility(reportA, reportB)`.
    5.  Return the resulting analysis JSON.

### 3.5. AI Chat Service

* **Responsibility:** Manages real-time chat via WebSockets.
* **API Endpoint (Gin):**
    * `GET /api/v1/chat`: Handles the HTTP Upgrade request.
* **Implementation Logic:**
    1.  Use the `gorilla/websocket` library. The handler will upgrade the HTTP connection.
    2.  After upgrading, enter a `for` loop to continuously `ReadMessage` from the client's WebSocket connection.
    3.  For each message, construct a prompt for the LLM API, including the user's message and contextual data about the current match analysis (retrieved from a short-lived Redis key like `chat_context:<user_id>`).
    4.  Make a streaming HTTP request to the LLM. As response chunks arrive, immediately use `WriteMessage` to send them to the client's WebSocket, providing a real-time streaming experience.

---

## 4. Asynchronous Communication (Future)

* **Technology:** Apache Kafka
* **Topic:** `user_profile_updates`
* **Event Schema:** When a user updates their profile, the User Service will publish an event to Kafka.
    ```json
    {
      "event_id": "uuid-v4",
      "event_type": "USER_PROFILE_UPDATED",
      "timestamp": "iso_8601_timestamp",
      "payload": {
        "user_id": 123,
        "changed_fields": ["location", "interests"]
      }
    }
    ```
* **Consumer:** A future batch processing job (Spark) for generating recommendations will consume from this topic to get the latest user data.

---

## 5. Deployment & Infrastructure (Kubernetes)

* **Containerization:** Each microservice will be packaged as a Docker image.
* **Orchestration:** Kubernetes will be used to manage and scale the services.
* **Key Kubernetes Objects:**
    * **Deployments:** For each service, defining the desired number of replicas and container image.
    * **Services:** To provide stable internal network endpoints for each microservice (e.g., `http://user-service`).
    * **Ingress:** An Ingress controller (like NGINX or Traefik) will manage external traffic from the internet, routing requests to the API Gateway. It will also handle TLS termination.
    * **ConfigMaps & Secrets:** To manage configuration and sensitive credentials (API keys, DB passwords) separately from the application code.

