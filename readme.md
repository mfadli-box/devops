# 🚀 ict_monorepo devops stack

Integrated documentation for the architecture management, development, and deployment of all ICT services.

---

## 📌 1. Folder Structure & Project Architecture

This project employs a **Monorepo** approach, managed as an integrated DevOps stack using Docker Compose.

```text
.
├── agents.md                      # Universal AI Agent & Assistant Guide
├── readme.md                      # Project Documentation (This File)
├── docker-compose.yml             # Core Stack Orchestration
├── ict_base/                      # Database Layer & Prisma ORM
│   └── prisma/
│       └── schema/                # 1 .prisma file per Database Cluster
├── ict_auto/                      # Automation Services (Go CLI)
│   ├── ict_nginx_log/
│   │   ├── main.go
│   │   └── Dockerfile
├── ict_rotate_log/
│   │   ├── main.go
│   │   └── Dockerfile
│   └── [other_automation_service]/
│       ├── main.go
│       └── Dockerfile
├── ict_rest/                      # Backend REST API (Gin Gonic)
│   ├── backbone/                  # Router, Database Config, Session Middleware
│   ├── skeleton/                  # API Modules (Linear Layer Flow Pattern)
│   ├── main.go
│   └── Dockerfile
└── ict_site/                      # Frontend Web Application (Next.js)
    ├── src/                       # Main Source Code (TypeScript)
    └── Dockerfile
```

---

## 🛠️ 2. Technology Stack Specifications

### 📁 ict_base (Database & ORM)
* **Database Engine:** PostgreSQL 18
* **ORM:** Prisma ORM 7.8.0
* **Schema Pattern:** Multi-file schema layout.
* **Location Rule:** Each `.prisma` file inside `ict_base/prisma/schema/` represents a single, independent database cluster.

### 📁 ict_auto (Automation)
* **Runtime:** Go 1.26.4
* **Application Pattern:** Single-file Go CLI applications.
* **Isolation Rule:** Each subdirectory within `ict_auto/` operates as an isolated automation microservice.

### 📁 ict_rest (Backend API)
* **Runtime & Framework:** Go 1.26.4 & Gin Gonic
* **Architecture:** Each cluster directory inside `skeleton/` strictly mirrors one database schema file from `ict_base`.
* **Data Flow Pattern:** Must strictly follow the linear layer execution: `Template (Struct & Interface) -> Repository -> Usecase -> Handler`.

### 📁 ict_site (Frontend Web)
* **Framework:** Next.js 16.2.9 (App Router) & React 19.2.4
* **Language:** TypeScript 5 (Strict Mode)
* **Design & UI:** Tailwind CSS 4 & shadcn/ui

---

## 🐋 3. DevOps & Dockerization Guide

Each application component must include a multi-stage `Dockerfile` to ensure production image sizes remain minimal, optimized, and secure.

### Docker Compose Orchestration
The entire stack is managed centrally through a unified `docker-compose.yml` file located at the project root with the following standards:
* **Data Persistence:** Uses Named Volumes to ensure PostgreSQL 18 data is safely persisted.
* **Network Isolation:** Uses a dedicated custom internal network for secure inter-service communication.

### Basic Stack Execution Commands

Run or rebuild the entire environment locally in detached mode:
```bash
docker compose up -d --build
```

Stop and remove all running containers and networks:
```bash
docker compose down
```

---

## 🌐 4. Development Workflow

1. **Database & API Synchronization:** When adding a new feature, define the database schema in `ict_base/prisma/schema/[cluster_name].prisma` first, then create its handling module inside `ict_rest/skeleton/[cluster_name]/`.
2. **Automation Development:** Add a dedicated new folder under `ict_auto/` for every new cron job, worker, or script function.

---

## 🚀 5. Publish Stack

Follow these steps sequentially to provision and deploy the stack into production.

### Step 1: Create Custom Isolated Network
```bash
docker network create --attachable --driver=bridge --subnet=172.99.66.0/24 --gateway=172.99.66.254 blackbox
```

### Step 2: Clone Repository & Provision Directories
```bash
git clone https://github.com/mfadli-box/devops.git
cd devops
git pull origin main

# Create persistent storage directories
mkdir -p ict_docs/pgsql
mkdir -p ict_docs/pgadmin
mkdir -p ict_docs/pgbackup

# Setup user mapping for pgbouncer
vi ict_docs/pgbouncer.txt
# Insert inside pgbouncer.txt:
# "postgres" "rahasia"
# "dbe" "rahasia"

# Apply global execution permissions to docs folder
chmod -R 777 ict_docs
```

### Step 3: Populate Global Environment Variables
```bash
vi .env
```
```env
PG_HOST=ict_base
PG_POOL=ict_pool
PG_PORT=5432
PG_DATA=ict
PG_USER=dbe
PG_PASS=rahasia
AD_EMAIL=admin@localhost
AD_PASS=rahasia
IS_POOL=true
ES_LINK=http://elasticsearch:9200
BE_POOL=http://ict_rest:36665
FT_HTTP=150
RE_PATH=archive
RE_NORMAL=7
RE_ATTACK=90
```

### Step 4: Pull Latest Core Artifacts & Trigger Production Build
```bash
git pull origin main
docker compose up -d --build

# Initialize Database Schema & Client Generation
cd ict_base
npx prisma generate
npx prisma migrate dev --name init

# Restart core stack to pick up initial migrations
cd ..
docker compose down
docker compose up -d
```

### For Optimize postgresql Change docker-compose

```
services:
  ict_base:
    image: postgres:18-alpine
    container_name: ict_base
    shm_size: 10gb
    deploy:
      resources:
        limits:
          cpus: '14.0'
          memory: 10g
    command: >
      postgres 
      -c shared_buffers=8GB 
      -c effective_cache_size=8GB 
      -c work_mem=16MB 
      -c maintenance_work_mem=2GB 
      -c max_connections=150 
      -c checkpoint_completion_target=0.9 
      -c wal_buffers=16MB 
      -c random_page_cost=1.1 
      -c effective_io_concurrency=200
```
