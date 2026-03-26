# Database Commands Guide

Here is a list of all the useful commands you might need to manage your unified database setup (PostgreSQL and MongoDB) running in Docker.

## 🐳 Docker Management Commands
Run these commands from your host machine (Windows CMD, PowerShell, or terminal) inside the `e:\Database` folder.

- **Start all databases in the background:**
  ```bash
  docker-compose up -d
  ```

- **Stop all databases (Keep data):**
  ```bash
  docker-compose down
  ```

- **Stop all databases AND wipe all data:** *(Warning: Destructive)*
  ```bash
  docker-compose down -v
  ```

- **View live logs of all databases:**
  ```bash
  docker-compose logs -f
  ```

- **Check status of your database containers:**
  ```bash
  docker ps
  ```

---

## 🐘 PostgreSQL Commands

### 1. Connecting to PostgreSQL
To execute database commands, you first need to log into the Postgres container shell.

- **Log in as `admin` to the `unified_db` database:**
  ```bash
  docker exec -it unified_postgres psql -U admin -d unified_db
  ```
  *(Your prompt will change to `unified_db=#` when successful)*

### 2. SQL Commands (Inside PostgreSQL Shell)
Once you are logged into the shell (from the command above), you can run these SQL commands:

- **List all databases:**
  ```sql
  \l
  ```

- **Connect to a different database:**
  ```sql
  \c database_name
  ```

- **Create a new database:**
  ```sql
  CREATE DATABASE new_app_db;
  ```

- **List all tables in the current database:**
  ```sql
  \dt
  ```

- **Exit the PostgreSQL shell:**
  ```sql
  \q
  ```

---

## 🍃 MongoDB Commands

### 1. Connecting to MongoDB
To execute MongoDB commands, you first need to log into the Mongo container shell (mongosh).

- **Log in as `admin`:**
  ```bash
  docker exec -it unified_mongodb mongosh -u admin -p 8fKx9Pq2LmZ4vW7y --authenticationDatabase admin
  ```
  *(Your prompt will change when successful)*

### 2. Mongo Commands (Inside mongosh Shell)
Once you are logged into the mongosh shell, you can run these commands:

- **Show all databases:**
  ```javascript
  show dbs
  ```

- **Switch to a database (or create it if it doesn't exist):**
  ```javascript
  use new_app_db
  ```

- **Show all collections (like tables) in the current database:**
  ```javascript
  show collections
  ```

- **Exit the MongoDB shell:**
  ```javascript
  exit
  ```

---

## 🟥 Redis Commands

### 1. Connecting to Redis
To execute Redis commands, log into the Redis container shell (redis-cli).

- **Log in with the password:**
  ```bash
  docker exec -it unified_redis redis-cli -a p9Kj2mT7vWcD4s8X
  ```
  *(Your prompt will change to `127.0.0.1:6379>`)*

### 2. Redis Commands (Inside redis-cli)
Once you are logged in, you can run these basic commands:

- **Test if Redis is working:**
  ```redis
  PING
  ```
  *(It should reply `PONG`)*

- **Set a simple key-value and retrieve it:**
  ```redis
  SET mykey "Hello World"
  GET mykey
  ```

- **Exit the Redis shell:**
  ```redis
  exit
  ```

