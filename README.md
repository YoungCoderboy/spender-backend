# 💰 Spender Backend

A Go-based expense tracking API that uses **Gemini AI** to parse natural language text (like SMS or chat) into structured financial transactions.

## 🚀 Features

* **AI Parsing**: Converts "Spent 25 USD on coffee" into structured JSON using Gemini 1.5 Flash.
* **Asynchronous Processing**: Uses Redis and a worker pattern to handle AI requests without slowing down the API.
* **Secure Auth**: JWT-based authentication and Bcrypt password hashing.
* **Infrastructure as Code**: Fully containerized Postgres and Redis setup.

---

## 🛠️ Prerequisites

* [Go 1.21+](https://go.dev/)
* [Docker & Docker Compose](https://www.docker.com/)
* [Gemini API Key](https://aistudio.google.com/)

---

## ⚙️ Setup & Installation

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/spender-backend.git
cd spender-backend

```

### 2. Configure Environment Variables

Create a `.env` file in the root directory:

```bash
# Database & Redis
DATABASE_URL=postgres://postgres:postgres@localhost:5432/spender?sslmode=disable
REDIS_ADDR=localhost:6379

# Secrets
JWT_SECRET=your_random_secret_string
GEMINI_API_KEY=your_google_api_key_here

```

### 3. Start Infrastructure (Databases)

```bash
docker-compose up -d postgres redis migrate

```

### 4. Run the Application

```bash
go mod tidy
go run cmd/apis/main.go

```

---

## 📡 API Endpoints

| Method | Endpoint | Description | Auth Required |
| --- | --- | --- | --- |
| `POST` | `/register` | Create a new account | No |
| `POST` | `/login` | Get JWT access token | No |
| `POST` | `/api/notify` | Send raw text for AI parsing | **Yes** |

### Example Request (`/api/notify`)

```bash
curl -X POST http://localhost:8080/api/notify \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{"rawText": "Dinner at Five Guys for 50 AED"}'

```

---

## 📂 Project Structure

* `cmd/` - Application entry points.
* `internal/domain/` - Business logic interfaces and models.
* `internal/service/` - Core logic (Auth, Engine, AI integration).
* `internal/repository/` - Database and Redis implementations.
* `migration/` - SQL schema files.

---

## 🧪 Testing

To verify the database tables were created:

```bash
docker exec -it spender_db psql -U postgres -d spender -c "\dt"

```
