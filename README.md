# I/O Linux Server

![Logo](react/src/assets/logo.png)

**I/O Linux Server** is a modern dashboard for managing your Linux system using native tools.  
It aims to unify essential functionality in a single web-based interface without reinventing the wheel.

---

## 🧠 Philosophy

Most Linux distributions already come with powerful tools for monitoring and control — `top`, `systemctl`, `ss`, etc.  
This project is about **leveraging those existing tools** by exposing their input/output via a friendly, minimal, and customizable web UI.  

Instead of replacing the Linux experience, **I/O Linux Server visualizes it.**

---

## ⚙️ Stack

- **Frontend:** React (Vite + MUI - based on [Mira Pro theme](https://mira.bootlab.io/))  
- **Backend:** Go + Air (for development)
- **Go Rest API:** Gin
- **Go Websocket:** gorilla

---

## 🚀 Features

- 🖥️ System stats dashboard: CPU, memory, disk, network
- 🧠 Process viewer: see running processes live
- 💻 Terminal output: view real-time output of Linux commands
- 🔐 Authentication via PAM (or other pluggable systems)
- 🧱 Static frontend serving in production
- 🛡️ WireGuard management UI
- 🐳 Docker Compose manager

---

## 📦 Getting Started

### Install dependencies

**For Debian/Ubuntu:**

```bash

sudo apt update # Update package list
sudo apt install -y make curl git lm-sensors libpam0g-dev git # Install required tools
```

**For Fedora / RHEL / CentOS:**

```bash
sudo dnf install -y make curl git lm_sensors pam-devel git # Install required tools
```

### Clone the repo

```bash
git clone https://github.com/mordilloSan/IO-Linux-Server
cd IO-Linux-Server
```

## 🛠️ Available Commands

This repo uses make to simplify standard operations.

```bash
make setup            # Install Node.js, Go (if missing) and frontend deps
make test             # Run frontend lint + type checks
make build            # Run full build (frontend + backend)
make build-frontend   # Build Vite React app
make build-backend    # Compile Go backend with version metadata
make dev              # Start frontend (Vite) and backend (Go) in dev mode
make prod             # Build react production files and run production backend
make binary           # Compile Go backend and run binary
make clean            # Remove build artifacts
make check-env        # Verify .env and required variables

```

✅ Both `make dev` and `make prod` will run all necessary setup automatically.

---

## 🔐 Logging In

This project uses **PAM authentication** to log in directly to your Linux system using your own username and password.

---

## 👨‍💼 Development Workflow

The development environment is fully set up with a **hot-reloading backend** (Go + Gin) and a **fast-refresh frontend** (Vite + React).

### 📆 Start Development

```bash
make dev
```

This will:

1. Start the **Go backend** using [Air] — any code changes automatically rebuild and restart the server.
2. Start the **React frontend** using Vite’s dev server with HMR (Hot Module Replacement).
3. Proxy frontend API requests to the backend, so everything just works.

---

### 🔪 How It Works

Under the hood:

- **Air** watches Go files and rebuilds the backend on changes.
- The **Air config** lives in `go-backend/.air.toml`.
- The **React frontend** runs in `react/` and talks to the backend via Vite's proxy (see `vite.config.ts`).
- **Makefile** handles all orchestration — use `make dev` as your single command to launch both.

💡 Tip: You can edit `.env` files for dev-specific settings (like ports, proxy targets, etc.).

---

## 📁 Project Structure

```
IO_Linux_Server/
├── go-backend/       # Gin-powered backend
├── react/            # Vite + React frontend
├── .env              # Environment variables
├── makefile          # Automation of builds & setup
└── README.md         # You're reading it!
```

---

## 📃 License

MIT License — feel free to use, fork, or contribute!

---

## 🙋‍♂️ Author

Created by [@mordilloSan](https://github.com/mordilloSan)  
📧 miguelgalizamariz@gmail.com  

