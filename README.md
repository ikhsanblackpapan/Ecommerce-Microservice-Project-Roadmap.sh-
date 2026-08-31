# 🛒 Microservices E-Commerce System

Sistem backend e-commerce berbasis arsitektur **Microservices** modern dengan pendekatan **Polyglot Architecture**, **Event-Driven Messaging**, dan **Containerization**.

---

## 🏛️ System Architecture

```text
                     [ CLIENT / POSTMAN ]
                               │
                               ▼
                   [ NGINX API Gateway :8000 ]
                               │
   ┌───────────────────┬───────┴───────────┬───────────────────┐
   │                   │                   │                   │
   ▼                   ▼                   ▼                   ▼
[ User Service ]  [ Product Service ] [ Order Service ]   [ Payment Service ]
 (Node.js +         (Laravel + PHP)    (Node.js +          (Node.js +
  Prisma)            :8000              Prisma)             Prisma)
  :8001                                 :8004               :8005
                                           │                   │
                                           │ (Publish Event)   │ (HTTP Update)
                                           ▼                   │
                                     [ RabbitMQ :5672 ]        │
                                           │                   │
                                           │ (Consume Event)   │
                                           ▼                   │
                                     [ Cart Service ] ◄────────┘
                                      (Go / Golang + 
                                       Redis)
                                       :8003


Fitur Utama & Keunggulan Arsitektur
Polyglot Architecture: Menggabungkan keunggulan Node.js, PHP (Laravel), dan Go (Golang) dalam satu ekosistem.

Event-Driven Checkout: order-service mempublikasikan event ORDER_CREATED ke RabbitMQ, yang kemudian didengarkan oleh cart-service untuk mengosongkan keranjang di Redis secara asynchronous.

High-Performance Caching: Menggunakan Redis pada cart-service untuk operasi pembacaan & penulisan keranjang belanja yang cepat.

Centralized Security: Autentikasi terpusat berbasis JWT (JSON Web Token) lewat API Gateway.

Full Containerization: Semua service & database di-orchestrate menggunakan Docker Compose.