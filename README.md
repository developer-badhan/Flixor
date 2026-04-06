# 🎬 Flixor — AI-Powered Movie Streaming & Discovery Platform

> Smart movie discovery meets scalable backend engineering.

Flixor is an AI-powered movie discovery and streaming platform built with **Gin (Go), React, and MongoDB**. It allows users to explore movies from **legal public sources**, stream content seamlessly, and receive intelligent recommendations based on their preferences. The project focuses on **clean backend architecture, scalable design, and real-world full-stack development practices**.

---

# 📜 **License**: MIT License

---

# 🚀 Why Flixor?

This project is designed to demonstrate:

* Real-world backend architecture
* Scalable frontend structure
* AI integration in modern apps
* Clean separation of concerns
* Follows Test-Driven-Development

---

# 🧠 Core Features

## 🔐 Authentication

* JWT-based authentication
* Password hashing (bcrypt)
* Protected routes (middleware)
* OTP email verification (extendable)

## 🎬 Movie System

* Fetch movies from legal sources (Internet Archive)
* Store metadata in MongoDB
* Movie detail APIs

## ▶️ Streaming

* Stream via public video URLs
* React video player
* View count tracking

## 🔍 Search & Filtering

* Search by title
* Genre filtering
* Pagination (optimized queries)

## ❤️ User Features

* Watchlist
* Like / Dislike
* Watch history

## 🤖 AI Recommendation

* Rule-based (genre + history)
* Gemini-powered recommendations
* Natural language movie search

## 📊 Analytics

* Trending movies
* Most watched

---


# 🏗️ Backend Architecture (Gin)

> Follows **clean architecture pattern**

```bash
backend/
│── cmd/
│   └── main.go              # Entry point
│
│── config/                  # Env & DB config
│
│── internal/
│   ├── handler/             # HTTP layer (controllers)
│   ├── service/             # Business logic
│   ├── repository/          # DB access layer
│   ├── model/               # Structs (schemas)
│   ├── middleware/          # JWT, logging
│   └── router/              # Route definitions
│
│── pkg/                     # Reusable utilities
│   ├── utils/
│   └── validator/
```

---

## 🧠 Backend Thinking

```text
Request → Router → Middleware → Handler → Service → Repository → DB
```

---

# 🎨 Frontend Architecture (React)

> Follows **feature-based + scalable structure**

```bash
frontend/
│── src/
│   ├── app/                 # App config (routes, providers)
│   ├── components/          # Reusable UI components
│   ├── features/            # Feature-based modules
│   │   ├── auth/
│   │   ├── movies/
│   │   ├── recommendation/
│   │   └── user/
│   ├── pages/               # Route-level pages
│   ├── services/            # API calls
│   ├── hooks/               # Custom hooks
│   ├── utils/               # Helper functions
│   └── assets/              # Images, styles
```

---

## 🧠 Frontend Thinking

```text
Page → Feature → Component → API Service → Backend
```


# 🛠️ Tech Stack

**Backend**
```
* Go + Gin Framework
* MongoDB (Database: NoSQL)
```

**Frontend**
```
* Vite + React.js (Library)
```

# 🤖 AI
```
Gemini API (free tier)
```

# 🔄 Data Flow

```text
React UI → API Service → Gin Backend → MongoDB → AI → Response
```

---

# 🔒 Security

* JWT authentication
* Middleware protection
* Input validation

---

# ⚠️ Legal Note

* Only public-domain / legal movie sources are used
* No copyrighted content is hosted

---

# ⚙️ Setup & Run Locally

---
📦 Prerequisites
Node.js (v18+)
Go (v1.20+)
MongoDB (local or Atlas)


## 🔑 Environment Variables

Create a `.env` file in the backend root:

```env
PORT=5000
MONGO_URI=mongodb://localhost:27017/flixor
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRES_IN=24h
EMAIL_SERVICE=gmail
EMAIL_USER=your_email@gmail.com
EMAIL_PASS=your_app_password
GEMINI_API_KEY=your_gemini_api_key
MOVIE_API_BASE_URL=https://archive.org/advancedsearch.php
```

## 🔧 Backend Setup

```bash
git clone https://github.com/developer-badhan/Flixor.git
cd backend
go mod tidy
go run cmd/main.go
```

---

## 🎨 Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

---

## 🏗️ Production Build

### Frontend
```bash
npm run build
```

### Backend
```bash
go build -o app cmd/main.go
./app
```

---

## 🌐 Access URLs

- Frontend: http://localhost:5173
- Backend: http://localhost:5000

---

# 🚀 Future Enhancements

* Advanced ML recommendations
* Microservices architecture
* Docker deployment
* CI/CD pipeline

---



# 💡 What This Project Shows

* Backend architecture skills
* API design thinking
* Full-stack integration
* AI-powered features

---

