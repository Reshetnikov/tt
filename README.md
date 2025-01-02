# Time Tracker

Explore the production version of the project:
[Time Tracker](http://time-tracker.me/)

## 🕒 Project Description

Time Tracker is a web application for tracking and managing time. It helps efficiently plan and control work processes.

## 🛠 Technologies

- **Language**: Go
- **Database**: PostgreSQL
- **Cache**: Redis
- **Email Service**: AWS SES or Mailgun
- **Hosting**: AWS

## Quick Start

### Prerequisites

- Docker
- Docker Compose

### Local Development

1. Clone the repository:

   ```bash
   git clone https://github.com/Reshetnikov/tt.git
   cd tt
   ```

2. Prepare the environment:

   ```bash
   # Copy the example configuration
   cp .env.example .env

   # Edit .env file with your settings
   ```

3. Run via Docker:

   ```bash
   docker-compose -f docker-compose.dev.yml up --build
   ```

4. Open in browser:
   http://localhost:8080
