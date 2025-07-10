# Matchmaker

Matchmaker is an astrology-based compatibility platform built as a suite of Go microservices. This repository contains high-level (HLD) and low-level (LLD) design documents as well as instructions for building and running the services locally.

## Architecture Overview

The application is composed of several independent services communicating over HTTP:

- **API Gateway** – Routes all client requests and verifies JWTs.
- **Auth Service** – Uses Google OAuth 2.0 for login and issues internal JWTs.
- **User Service** – Stores user profiles in PostgreSQL.
- **Astrology Report Service** – Retrieves and caches astrology reports from an external engine using MongoDB/Redis.
- **Match Analysis Service** – Calculates compatibility scores between two users.
- **AI Chat Service** – WebSocket service that integrates with a third-party LLM provider for real-time chat.

See `HLD_Readme.md` and `LLD_Readme.md` for full architectural details and diagrams.

## Building the Services

Prerequisites: Go 1.20+, Docker, and Docker Compose.

Each service lives in `services/<name>` and provides a `Dockerfile`. To build the Go binaries directly:

```bash
# API Gateway
cd services/gateway
go build -o gateway .

# Auth Service
cd ../auth
go build -o auth .

# User Service
cd ../user
go build -o user .

# Astrology Report Service
cd ../report
go build -o report .

# Match Analysis Service
cd ../analysis
go build -o analysis .

# AI Chat Service
cd ../chat
go build -o chat .
```

To build Docker images for deployment:

```bash
cd services/gateway
docker build -t matchmaker-gateway .
cd ../auth
docker build -t matchmaker-auth .
cd ../user
docker build -t matchmaker-user .
cd ../report
docker build -t matchmaker-report .
cd ../analysis
docker build -t matchmaker-analysis .
cd ../chat
docker build -t matchmaker-chat .
```

## Running Locally with Docker Compose

After building the images, return to the repository root and start all services with:

```bash
docker-compose up
```

The API Gateway will be available at `http://localhost:8080`.

## Environment Variables

Services use the following environment variables. Create a `.env` file or export them in your shell before running Docker Compose.

| Variable | Description |
| -------- | ----------- |
| `POSTGRES_URL` | Connection string for the User Service database |
| `MONGO_URL` | MongoDB connection for the Astrology Report Service |
| `REDIS_URL` | Redis endpoint for caching and chat sessions |
| `GOOGLE_OAUTH_CLIENT_ID` | Client ID for Google login |
| `GOOGLE_OAUTH_CLIENT_SECRET` | Client secret for Google login |
| `JWT_PRIVATE_KEY` | PEM-encoded RSA key used to sign JWTs |
| `ASTROLOGY_ENGINE_URL` | Endpoint of the external astrology engine |
| `ASTROLOGY_ENGINE_API_KEY` | API key for the astrology engine |
| `LLM_API_KEY` | API key for the chat service's LLM provider |
| `GOOGLE_OAUTH_REDIRECT_URL` | OAuth callback URL used by the Auth Service |
| `USER_SERVICE_URL` | Endpoint of the User Service |
| `REPORT_SERVICE_URL` | Endpoint of the Astrology Report Service |
| `AUTH_SERVICE_URL` | Endpoint of the Auth Service |
| `MATCH_SERVICE_URL` | Endpoint of the Match Analysis Service |
| `CHAT_SERVICE_URL` | Endpoint of the AI Chat Service |
| `LLM_API_URL` | Endpoint of the LLM provider for the Chat Service |

## Third-Party Dependencies

- **Google OAuth** – Handles user authentication. Configure `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET`.
- **PostgreSQL** – Stores user profiles. Set `POSTGRES_URL` appropriately.
- **MongoDB** – Document store for cached astrology reports via `MONGO_URL`.
- **Redis** – Used for caching and chat sessions through `REDIS_URL`.
- **External Astrology Engine** – Generates birth chart reports; requires `ASTROLOGY_ENGINE_URL` and `ASTROLOGY_ENGINE_API_KEY`.
- **LLM Provider** – Supplies responses for the AI chat feature and needs `LLM_API_KEY`.

## API Usage Examples

### Start Google Login

```http
GET /api/v1/auth/google/login
```

### OAuth Callback

```http
GET /api/v1/auth/google/callback?code=<auth_code>
```

### Retrieve Current User

```http
GET /api/v1/users/me
Authorization: Bearer <jwt>
```

### Update Profile

```http
PUT /api/v1/users/me
Authorization: Bearer <jwt>
Content-Type: application/json

{
  "location": "San Francisco",
  "photoURL": "https://example.com/me.jpg"
}
```

### Request Match Analysis

```http
POST /api/v1/analysis
Authorization: Bearer <jwt>
Content-Type: application/json

{
  "personA": {"dob": "1990-01-01", "tob": "12:00:00", "lat": 40.71, "lon": -74.00},
  "personB": {"dob": "1992-02-02", "tob": "06:30:00", "lat": 34.05, "lon": -118.24}
}
```

### AI Chat via WebSocket

```http
GET /api/v1/chat
```

After upgrading the connection, send chat messages and stream the LLM responses.

---

Consult the HLD and LLD documents for detailed design decisions and diagrams.
