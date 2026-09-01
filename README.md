# 🛒 Microservices E-Commerce System

> Project Roadmap URL: https://roadmap.sh/projects/scalable-ecommerce-platform

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
  Prisma)            :8002              Prisma)             Prisma)
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



## ⚡ Cara Menjalankan Project

### Prerequisites
Pastikan perangkat kamu sudah ter-install software berikut:
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (Sudah dalam kondisi berjalan/active)
- [Git](https://git-scm.com/)

---

### Langkah-Langkah Instalasi & Pengujian

1. **Clone Repository**
   ```bash
   git clone [https://github.com/ikhsanblackpapan/Ecommerce-Microservice-Project-Roadmap.sh-.git](https://github.com/ikhsanblackpapan/Ecommerce-Microservice-Project-Roadmap.sh-.git)
   cd Ecommerce-Microservice-Project-Roadmap.sh-




1. Setup Environment Variables (.env)
Buat/sesuaikan file .env di masing-masing service (user-service, order-service, payment-service, dll.) dengan merujuk pada file .env.example yang tersedia di tiap folder service.

2. Jalankan Seluruh Ecosystem dengan Docker Compose
Jalankan perintah ini di root folder project untuk build dan menjalankan semua container service, database, Redis, RabbitMQ, serta NGINX API Gateway:

docker compose up -d --build


Jalankan Migrasi Database (Prisma ORM)
Eksekusi skema database ke container MySQL/MariaDB yang sedang berjalan:


# Migrasi schema User Service
docker compose exec user-service npx prisma db push

# Migrasi schema Order Service
docker compose exec order-service npx prisma db push

# Migrasi schema Payment Service
docker compose exec payment-service npx prisma db push

Full Containerization: Semua service & database di-orchestrate menggunakan Docker Compose.
