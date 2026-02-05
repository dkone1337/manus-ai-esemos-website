# ESEMOS Go App

Eine vollständig selbst hostbare Go-Anwendung mit Admin-Panel, inspiriert vom Design von esemos.de.

## Features
- **Dunkles Design:** Teal/Cyan Akzente, Glassmorphism-Effekte.
- **Admin Panel:** Verwaltung von Blog-Beiträgen (CRUD).
- **Go Backend:** Schnell, sicher und einfach zu deployen.
- **SQLite Datenbank:** Keine externe Datenbank nötig.

## Installation auf openSUSE Tumbleweed

### 1. Voraussetzungen
```bash
sudo zypper install go1.22 nginx
```

### 2. Kompilieren
```bash
go build -o esemos-app cmd/main.go
```

### 3. Systemd Service einrichten
```bash
sudo cp configs/esemos.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now esemos
```

### 4. Nginx konfigurieren
```bash
sudo cp configs/nginx.conf /etc/nginx/vhosts.d/esemos.conf
sudo systemctl reload nginx
```

## Admin Login
- **URL:** `/login`
- **Standard-User:** `admin`
- **Standard-Passwort:** `admin123` (Bitte nach dem ersten Login ändern!)
