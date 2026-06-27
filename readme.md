# 🚀 ICT Monorepo DevOps Stack

Integrated documentation for the architecture management, development, and deployment of all ICT services.

------------------------------

## 📌 1. Folder Structure & Project Architecture

This project employs a **Monorepo** approach, managed as an integrated DevOps stack using Docker Compose.

```text
.
├── .github/
│   └── copilot-instructions.md    # AI Copilot Guide
├── ict_base/                      # Database Layer & Prisma ORM
│   └── prisma/
│       └── schema/                # 1 .prisma file per Database Cluster
├── ict_auto/                      # Automation Service (Go CLI)
│   ├── ict_nginx_log/
│   │   ├── main.go
│   │   └── Dockerfile
│   └── ict_rotate_log/
│       ├── main.go
│       └── Dockerfile
├── ict_rest/                      # Backend REST API (Gin Gonic)
│   ├── backbone/                  # Router, Database Configuration, Session Middleware
│   ├── skeleton/                  # API Module (Paired 1:1 with DB Cluster)
│   ├── main.go
│   └── Dockerfile
├── ict_site/                      # Frontend Web Application (Next.js)
│   ├── src/                       # Main Source Code (TypeScript)
│   └── Dockerfile
├── docker-compose.yml             # Core Stack Orchestration
└── readme.md                      # Project Documentation (This File)
```

------------------------------

## 🛠️ 2. Technology Stack Specifications

### 📁 ict_base (Database & ORM)

* Database Engine: PostgreSQL 18
* ORM: Prisma ORM 7.8.0
* Schema: Multi-file schema. Each ".prisma" file in "ict_base/prisma/schema/" represents a single independent database cluster.

### 📁 ict_auto (Automation)

* Runtime: Go 1.26.4
* Application Pattern: Single-file Go CLI applications. Each subfolder is an isolated automation microservice.

### 📁 ict_rest (Backend API)

* Runtime & Framework: Go 1.26.4 & Gin Gonic
* Architecture: Each cluster folder inside "skeleton/" strictly mirrors 1 schema file in "ict_base" with a strict data flow pattern: {Template (Struct & Interface)} > {Repository} > {Usecase} > {Handler}

### 📁 ict_site (Frontend Web)

* Framework: Next.js 16.2.9 (App Router) & React 19.2.4
* Language & Design Language: TypeScript 5 & Tailwind CSS 4

------------------------------

## 🐋 3. DevOps & Dockerization Guide

Each application component must include a Multi-stage Dockerfile to ensure production image sizes remain minimal and secure.

### Docker Compose Orchestration

The entire stack is run centrally through a "docker-compose.yml" file in the root folder with the following standards:

* Uses Named Volumes to ensure PostgreSQL 18 data is persistent.
* Uses a dedicated internal network (Custom Network) for secure inter-service communication.

### Basic Stack Execution Commands

Run the entire environment locally:

```
docker-compose up -d --build
```

Stop all services:

```
docker-compose down
```

------------------------------

## 🌐 4. Development Workflow

   1. Database & API Synchronization: When adding a new feature, define the database schema in "ict_base/prisma/schema/[cluster_name].prisma" first, then create its handling module in "ict_rest/skeleton/[cluster_name]/".

   2. Automation Development: Add a new folder under "ict_auto/" if you want to create a cron or worker function.

------------------------------

## 🚀 5. Publish Stack

Copy the git clone command for this repository.

### Create Network First

```
docker network create --attachable --driver=bridge --subnet=172.99.66.0/24 --gateway=172.99.66.254 blackbox
```

### Clone Git, Create Folder Inside Project Directory

```
git clone https://github.com/mfadli-box/devops.git
cd devops
git pull origin main
mkdir ict_docs/pgsql
mkdir ict_docs/pgadmin
mkdir ict_docs/pgbackup

vi ict_docs/pgbouncer.txt

"postgres" "rahasia"
"dbe" "rahasia"


chmod -R 777 ict_docs
```

### Create File Environment

```
vi .env

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

### Get Latest update and Build up Docker

```
git pull origin main
docker-compose up -d --build

cd erp_base
npx prisma generate
npx prisma migrate dev --name init

# docker-compose down
```

### For Optimize postgresql Change docker-compose

```
services:
  ict_base:
    shm_size: 10gb
    mem_limit: 10g
    cpus: 14
    command:
      - "postgres"
      - "-c"
      - "shared_buffers=8GB"
      - "-c"
      - "effective_cache_size=8GB"
      - "-c"
      - "work_mem=16MB"
      - "-c"
      - "maintenance_work_mem=2GB"
      - "-c"
      - "max_connections=150"
      - "-c"
      - "checkpoint_completion_target=0.9"
      - "-c"
      - "wal_buffers=16MB"
      - "-c"
      - "random_page_cost=1.1"
      - "-c"
      - "effective_io_concurrency=200"
```

------------------------------
