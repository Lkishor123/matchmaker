# Astrology Matchmaking Platform - High-Level Design (HLD)

## 1. Introduction

This document outlines the high-level architecture for an astrology-based matchmaking platform. The primary goal of the Minimum Viable Product (MVP) is to provide users with an on-demand astrological compatibility analysis between two individuals. The system is designed with a service-oriented approach to facilitate future scaling and the addition of features like profile recommendations and real-time chat.

## 2. Core Requirements

### 2.1. Functional Requirements

* **User Profiles:** Users must register and provide mandatory birth details (Date, Time, Place), gender, and current location. Photos and interests are optional.

* **Authentication:** Users will authenticate using Google's OAuth 2.0 (OIDC) service.

* **On-Demand Match Analysis:** A user can input their birth details and the birth details of a potential partner to receive a detailed astrological compatibility analysis.

* **AI Astrologer Chat:** After an analysis is generated, users can interact with an AI-powered chatbot to ask follow-up questions about the match.

### 2.2. Non-Functional Requirements

| Requirement | MVP Target (1k DAU) | Scaled Target (100k+ DAU) | 
 | ----- | ----- | ----- | 
| **Scalability** | Support 1,000 Daily Active Users. | Support 100,000 to 1M+ Daily Active Users. | 
| **Latency** | Match Analysis: < 2s. AI Chat: < 2s. General API: < 1s. | Maintain the same low latency targets at scale. | 
| **Availability** | **Approximately 99% (allowing for up to 3.65 days of downtime per year), which is acceptable for an MVP with scheduled maintenance windows.** | **99.9%**, multi-AZ deployment with database replication. |
| **Consistency** | Not a primary concern for the MVP. | Eventual consistency is acceptable for future features. | 
| **Geography** | Single region deployment is acceptable. | Global user base, requiring low latency worldwide. | 

## 3. High-Level Architecture

The system will be built using a **Service-Oriented Architecture (SOA)**, which will evolve into a full **Microservices Architecture** at scale. This approach promotes separation of concerns, independent scaling, and fault isolation.

### 3.1. Core Components

* **Client:** A React-based Single Page Application (SPA) will be the primary user interface.

* **API Gateway:** A single entry point for all client requests. It acts as the public-facing entry point, terminating TLS connections and handling cross-cutting concerns like routing, JWT validation, rate limiting, and request aggregation before forwarding traffic to internal services.

* **Auth Service:** Manages the "Login with Google" OIDC flow and issues internal JWTs to authenticated users.

* **User Service:** Manages user profile data (birth details, location, etc.).

* **Astrology Report Service:** A caching wrapper around an external astrology engine. It generates and caches detailed astrological reports to avoid redundant, costly calculations.

* **Match Analysis Service:** The core logic engine that takes two astrological reports and computes their compatibility.

* **AI Chat Service:** Manages real-time, bidirectional communication with clients via WebSockets for the AI Astrologer feature.

### 3.2. Data Flow Diagram

```
+-----------+      +----------------+      +---------------+      +----------------+
|           |----->|                |----->|               |----->|                |
|  Client   |      |  API Gateway   |      |  Auth Service |      |  Google OAuth  |
|  (React)  |<-----| (Validates JWT)|<-----|  (Issues JWT) |<-----|                |
+-----------+      +-------+--------+      +---------------+      +----------------+
                         |
                         | (Routes Requests)
            +------------+------------------+
            |                               |
  +---------V---------+          +----------V----------+
  |                   |          |                     |
  |  Match Analysis   |          |    AI Chat Service  |
  |      Service      |          |    (WebSockets)     |
  +---------+---------+          +----------+----------+
            |                               |
            | calls                         | calls
  +---------V---------+          +----------V----------+
  |                   |          |      3rd Party      |
  | Astrology Report  |          |       LLM API       |
  |      Service      |          +---------------------+
  +---------+---------+
            |
            | calls
  +---------V---------+
  |    External       |
  | Astrology Engine  |
  +-------------------+
```

*Note: The User Service interacts with the PostgreSQL DB. The Astrology Report Service interacts with both Redis (L1 Cache) and MongoDB (L2 Cache/Persistent Store). The Match Analysis Service calls the Astrology Report Service but does not directly access databases.*

## 4. Data Storage Strategy

A **polyglot persistence** model will be used to select the optimal database technology for each type of data.

* **User Database (PostgreSQL):**

  * **Reasoning:** User profile data is structured and relational. A SQL database provides strong consistency (ACID) and data integrity for core user accounts. GORM will be used as the ORM. While a NoSQL database could also store this data, a relational model is superior here for enforcing schema integrity and managing relationships between users and their core details, which are not expected to change structure frequently.

* **Astrology Report Database (MongoDB):**

  * **Reasoning:** The astrological reports are complex, nested JSON objects. A NoSQL Document DB like MongoDB is ideal for storing this semi-structured data without a rigid schema.

* **Cache (Redis):**

  * **Reasoning:** An in-memory key-value store like Redis will be used for two purposes:

    1. **L1 Cache:** For frequently accessed astrology reports to reduce latency before hitting MongoDB.

    2. **Session Cache:** To store temporary context for the AI Chat Service.

## 5. Scalability & Evolution Plan

The architecture is designed to evolve gracefully from the MVP to a large-scale global service.

| Aspect | MVP Strategy (1k DAU) | Scaled Strategy (100k+ DAU) | Rationale for Change | 
 | ----- | ----- | ----- | ----- | 
| **Architecture** | **Modular Monolith** on 1-2 servers. | **Full Microservices Architecture** on Kubernetes (EKS/GKE). | Independent scaling, deployment, and fault isolation are critical at scale. | 
| **Availability** | **Approximately 99% (allowing for up to 3.65 days of downtime per year), which is acceptable for an MVP with scheduled maintenance windows.** | **99.9%**, multi-AZ deployment with database replication. | Multi-AZ provides redundancy and automatic failover, eliminating single points of failure. | 
| **Geography** | Single region deployment (e.g., `us-east-1`). | **Global Deployment** in multiple regions (US, EU, APAC). | A **CDN** will serve static assets, and **GeoDNS** will route API traffic to the nearest region, minimizing latency for global users. | 
| **Async Communication** | Not required for MVP. | **Kafka** will be introduced for asynchronous events (e.g., `user_profile_updates`) to decouple services for future features. | Event-driven architecture improves resilience and scalability compared to synchronous inter-service calls. | 

## 6. Security Considerations

* **Authentication:** OIDC via Google, with JWTs for session management.

* **Authorization:** Role-Based Access Control (RBAC) will be implemented as features require different permission levels. This can be implemented by embedding a user's roles (e.g., 'user', 'moderator') directly into the JWT claims, allowing the API Gateway to perform fast authorization checks.

* **Encryption:**

  * **In Transit:** All communication will use TLS/HTTPS.

  * **At Rest:** Databases will use transparent data encryption (TDE).

* **Threat Mitigation:** Standard practices will be followed: prepared statements (via GORM) to prevent SQLi, output encoding to prevent XSS, and anti-CSRF tokens. A CDN and rate limiting at the API Gateway will mitigate DDoS attacks.

## 7. Technology Stack Summary

* **Frontend:** React

* **Backend:** Go

* **API Framework:** Gin

* **Relational DB:** PostgreSQL

* **ORM:** GORM

* **NoSQL DB:** MongoDB

* **Caching:** Redis

* **Messaging:** Apache Kafka (for future use)

* **Deployment:** Docker, Kubernetes
