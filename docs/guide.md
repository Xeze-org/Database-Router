# Guide: Connecting Multiple Apps to Unified Docker Databases

This guide explains how different applications can connect to your unified PostgreSQL and MongoDB databases, even though the databases are inside Docker containers.

## 1. Network Concepts

When you run databases in Docker, there are two common scenarios for how your apps connect:
*   **Scenario A: Apps running on your Host Machine** (e.g., when you run `npm start` or `python app.py` directly on your Windows PC).
*   **Scenario B: Apps running inside Docker Containers** (e.g., when you put your app inside Docker and add it to `docker-compose.yml`).

### How Ports Work

In your `docker-compose.yml`, you used `ports: "5432:5432"`. 
Docker "maps" the internal container port to your host machine's port. This creates a tunnel from your Windows PC directly into the container.

---

## 2. Scenario A: Your Apps Run Directly on Windows

When your web app (React/Node/Python/Java) is running directly on Windows, it acts like Docker is a remote server running on `localhost`. Since Docker exposed the ports, your app uses `localhost` to go through the tunnel into the container.

```mermaid
flowchart TD
    subgraph Windows Host Machine
        App1[Node.js App 1]
        App2[Python App 2]
    end

    subgraph Docker Environment
        PG[(PostgreSQL Container \n Port: 5432)]
        MONGO[(MongoDB Container \n Port: 27017)]
    end

    App1 -- "postgresql://admin:pwd@localhost:5432/app1_db" --> PG
    App2 -- "mongodb://admin:pwd@localhost:27017/app2_db" --> MONGO
```

**Connection URLs from Windows:**
*   **Postgres App 1:** `postgresql://admin:3rHb6NmA5jUc8Tg1@localhost:5432/app1_db`
*   **Postgres App 2:** `postgresql://admin:3rHb6NmA5jUc8Tg1@localhost:5432/app2_db`
*   **MongoDB App 1:** `mongodb://admin:8fKx9Pq2LmZ4vW7y@localhost:27017/app1_db?authSource=admin`

*(Note: In Postgres, you must first log in using an SQL client and manually create the `app1_db` and `app2_db` databases before your apps can connect. MongoDB creates them automatically when data is inserted).*

---

## 3. Scenario B: Your Apps Run Inside Docker

If you later decide to containerize your apps (putting your Node or Python code into a Dockerfile and adding them to the `docker-compose.yml`), they **cannot** use `localhost`. Inside a container, `localhost` points to the container itself, not your Windows machine or the database container!

Instead, Docker has an internal DNS. Containers on the same Docker network can talk to each other using their **service name** (the name you gave them in `docker-compose.yml`).

```mermaid
flowchart TD
    subgraph Docker Network (Isolated from Host)
        App1Node[App 1 Container \n Name: api_service_1]
        App2Py[App 2 Container \n Name: api_service_2]
        
        PG[(PostgreSQL Container \n Name: postgres)]
        MONGO[(MongoDB Container \n Name: mongodb)]
    end

    App1Node -- "postgresql://admin:pwd@postgres:5432/app1_db" --> PG
    App2Py -- "mongodb://admin:pwd@mongodb:27017/app2_db" --> MONGO
```

**Connection URLs inside Docker:**
*   **Postgres App 1:** `postgresql://admin:3rHb6NmA5jUc8Tg1@postgres:5432/app1_db`
*   *(Notice we use `postgres:` instead of `localhost:`)*
*   **MongoDB App 2:** `mongodb://admin:8fKx9Pq2LmZ4vW7y@mongodb:27017/app2_db?authSource=admin`

---

## 4. How to Create the Multiple Databases

### For MongoDB:
You don't need to do anything! Just write your connection string with the new database name. When your app tries to save data, MongoDB automatically creates the new database on the fly.

### For PostgreSQL:
Postgres requires the database to exist *before* the app can connect. Since you have 1 Postgres container but want many databases, you must create them.
1.  Open a database client on Windows (like DBeaver, pgAdmin, or VS Code SQLtools).
2.  Connect to `localhost` on port `5432` with username `admin` and password `3rHb6NmA5jUc8Tg1`.
3.  Run the SQL command: `CREATE DATABASE my_new_app_db;`
4.  Now your app can connect to `postgresql://admin:3rHb6NmA5jUc8Tg1@localhost:5432/my_new_app_db`
