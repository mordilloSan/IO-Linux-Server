# Linux I/O

![Logo](react/src/assets/logo.png)

**Linux I/O** is a modern dashboard for managing your Linux system using native tools.  
It aims to unify essential functionality in a single web-based interface without reinventing the wheel.

---

## 🧠 Philosophy

Most Linux distributions already come with powerful tools for monitoring and control — `docker`, `systemctl`, `nmcli`, etc.  
This project is about **leveraging those existing tools** by exposing their input/output via a friendly, minimal, and customizable web UI.  
As such we aim to rely on D-Bus connectivity, docker SDK and parsing linux commands. Hence the I/O meaning input/output

Instead of replacing the Linux experience, **Linux I/O visualizes it.**

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

sudo apt update
sudo apt install -y make curl git lm-sensors libpam0g-dev policykit-1
```

**For Fedora / RHEL / CentOS:**

```bash
sudo dnf install -y make curl git lm_sensors pam-devel dnf-plugins-core
```

### Clone the repo

```bash
git clone https://github.com/mordilloSan/IO-Linux-Server
cd IO-Linux-Server
```

---

## 🛠️ Available Commands

This repo uses `make` to simplify standard operations.

✅ Run `make` inside the project directory to view available commands  
✅ Both `make dev` and `make prod` will run all necessary setup automatically.  

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

---

## 🔐 Logging In

This project uses **PAM authentication** to log in directly to your Linux system using your own username and password.

---

## 👨‍💼 Development & Deployment Workflow

🔑 Secret File
For development and production (unless running the binary), edit the file called secret.env:

```env
SUDO_PASSWORD=your_password_here
```

This password is used for executing privileged operations via sudo.

### 🛠️ Development Mode

```bash
make dev
```

Runs Air for Go backend auto-reloads

Runs Vite dev server with proxying to Go API

Outputs all API paths and logs (from Gin)


### 🚀 Production Mode

```bash
make prod
```

- Compiles frontend via Vite

- Serves static assets using go run .

- No logging enabled by default


### 📦 Binary Mode

```bash
make binary
```

- Produces a compiled, self-contained Go binary

- Frontend is bundled inside

- Suitable for systemd and production deployment

- No .env or secret files needed after build

### 🔪 How It Works

Under the hood:

- **Air** watches Go files and rebuilds the backend on changes.
- The **Air config** lives in `go-backend/.air.toml`.
- The **React frontend** runs in `react/` and talks to the backend via Vite's proxy (see `vite.config.ts`).

💡 You can customize .env for ports, proxy settings, etc.

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

