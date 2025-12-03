<p align="center">
  <h1 align="center">PG Back Web</h1>
  <p align="center">
    <img align="center" width="70" src="https://raw.githubusercontent.com/eduardolat/pgbackweb/main/internal/view/static/images/logo.png"/>
  </p>
  <p align="center">
    🐘 Effortless database backups (PostgreSQL, ClickHouse) with a user-friendly web interface! 🌐💾
  </p>
</p>

<p align="center">
  <a href="https://github.com/eduardolat/pgbackweb/actions/workflows/ci.yaml?query=branch%3Amain">
    <img src="https://github.com/eduardolat/pgbackweb/actions/workflows/ci.yaml/badge.svg" alt="CI Status"/>
  </a>
  <a href="https://goreportcard.com/report/eduardolat/pgbackweb">
    <img src="https://goreportcard.com/badge/eduardolat/pgbackweb" alt="Go Report Card"/>
  </a>
  <a href="https://github.com/eduardolat/pgbackweb/releases/latest">
    <img src="https://img.shields.io/github/release/eduardolat/pgbackweb.svg" alt="Release Version"/>
  </a>
  <a href="https://hub.docker.com/r/eduardolat/pgbackweb">
    <img alt="Docker Pulls" src="https://img.shields.io/docker/pulls/eduardolat/pgbackweb"/>
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/eduardolat/pgbackweb.svg" alt="License"/>
  </a>
  <a href="https://github.com/eduardolat/pgbackweb">
    <img src="https://img.shields.io/github/stars/eduardolat/pgbackweb?style=flat&label=github+stars"/>
  </a>
</p>

> [!NOTE]
> **We're growing! New name, bigger future**
>
> PG Back Web is becoming **UFO Backup**! The new name reflects a future where the project expands beyond PostgreSQL, making powerful backups simple and accessible for everyone
>
> Curious about the roadmap or want to shape the project's future? Join the [community](https://ufobackup.uforg.dev/r/community) to discuss ideas and influence decisions, everyone's input is welcome!

## Why PG Back Web?

PG Back Web isn't just another backup tool. It's your trusted ally in ensuring the security and availability of your database data:

- 🎯 **Designed for everyone**: From individual developers to teams.
- ⏱️ **Save time**: Automate your backups and forget about manual tasks.
- ⚡ **Plug and play**: Don't waste time with complex configurations.
- 🔄 **Multi-database support**: Not just PostgreSQL - also supports ClickHouse and more coming soon!

## Features

### Core Capabilities

- 📦 **Intuitive web interface**: Manage your backups with ease, no database expertise required.
- 📅 **Scheduled backups**: Set it and forget it with cron-based scheduling. PG Back Web takes care of the rest.
- 📈 **Backup monitoring**: Visualize the status of your backups with detailed execution logs, file sizes, and execution history.
- 📤 **Instant download & restore**: Restore and download your backups when you need them, directly from the web interface. Supports restoring to any configured database with automatic version detection.
- 🔄 **Backup duplication**: Easily duplicate existing backup configurations to create new ones quickly.
- 👥 **Multi-user support**: Manage multiple users with session-based authentication.

### Database Support

- 🐘 **PostgreSQL**: Full support for PostgreSQL 13, 14, 15, 16, 17, and 18.
- 🚀 **ClickHouse**: Support for ClickHouse versions 22.8, 23.8, 24.1, and 24.3.
- 🔌 **Extensible architecture**: Easy to add support for additional database types.

### Storage Options

- 📁 **Local storage**: Store backups directly on the server filesystem.
- ☁️ **S3-compatible storage**: Support for AWS S3 and any S3-compatible storage (MinIO, DigitalOcean Spaces, etc.).
- 🔀 **Flexible destinations**: Configure multiple S3 destinations and choose per backup.
- 🔗 **Presigned URLs**: Secure, time-limited download links for S3-stored backups.

### Monitoring & Notifications

- ❤️‍🩹 **Health checks**: Automatically check the health of your databases and destinations.
- 🔔 **Webhooks**: Get notified via webhooks for:
  - Database health status changes (healthy/unhealthy)
  - Destination health status changes (healthy/unhealthy)
  - Backup execution success
  - Backup execution failures
- 📊 **Execution tracking**: Detailed logs for every backup execution with timestamps, file sizes, and status.

### Security & Reliability

- 🔒 **PGP encryption**: All sensitive data (connection strings, credentials) encrypted at rest using PostgreSQL PGP encryption.
- 🔐 **Password security**: Bcrypt hashing for user passwords.
- 🛡️ **Session management**: Secure session-based authentication with IP and user agent tracking.
- 🔑 **Encryption key**: Centralized encryption key management for all sensitive data.

### User Experience

- 🌚 **Dark mode**: Beautiful dark mode interface.
- 📱 **Responsive design**: Works seamlessly on desktop and mobile devices.
- ⚡ **Fast & lightweight**: Built with Go for performance and efficiency.
- 🎨 **Modern UI**: Clean, intuitive interface built with TailwindCSS and DaisyUI.

### Open Source

- 🛡️ **Open-source trust**: Open-source code under AGPL v3 license, backed by robust database tools (pg_dump, clickhouse-backup).
- 🔍 **Transparent**: Full source code available for review and contribution.

## Installation

PG Back Web is available as a Docker image. You just need to set 3 environment variables and you're good to go!

Here's an example of how you can run PG Back Web with Docker Compose, feel free to adapt it to your needs:

```yaml
services:
  pgbackweb:
    image: eduardolat/pgbackweb:latest
    ports:
      - "8085:8085" # Access the web interface at http://localhost:8085
    volumes:
      - ./backups:/backups # If you only use S3 destinations, you don't need this volume
    environment:
      # Optional environment variables are ignored, see the configuration section below for more details
      PBW_ENCRYPTION_KEY: "my_secret_key" # Change this to a strong key
      PBW_POSTGRES_CONN_STRING: "postgresql://postgres:password@postgres:5432/pgbackweb?sslmode=disable"
    depends_on:
      postgres:
        condition: service_healthy

  postgres:
    image: postgres:18
    environment:
      POSTGRES_USER: postgres
      POSTGRES_DB: pgbackweb
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - ./data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5
```

You can watch [this youtube video](https://www.youtube.com/watch?v=vf7SLrSO8sw) to see how easy it is to set up PG Back Web.

### Use Cases

PG Back Web is perfect for:

- **Development teams**: Centralized backup management for multiple databases
- **DevOps engineers**: Automated backup scheduling with monitoring and notifications
- **Small businesses**: Simple, reliable database backups without complex infrastructure
- **Multi-database environments**: Manage backups for PostgreSQL and ClickHouse from one interface
- **Compliance requirements**: Track all backup executions with detailed logs and history
- **Disaster recovery**: Quick restore capabilities with version-aware restoration

## Architecture

PG Back Web is built with a modular, extensible architecture:

- **Service layer**: Clean separation of concerns with domain-specific services (backups, databases, destinations, executions, restorations, webhooks, users, auth)
- **Database integration**: Pluggable database clients supporting multiple database types (PostgreSQL, ClickHouse)
- **Storage abstraction**: Unified storage interface supporting local filesystem and S3-compatible storage
- **Code generation**: SQLC-based query generation for type-safe database operations
- **Migration-first**: Database schema managed with Goose migrations
- **Cron scheduling**: Built-in cron scheduler for automated backup execution
- **Web framework**: Echo-based web framework with Alpine.js for interactive UI

## Configuration

You only need to configure the following environment variables:

- `PBW_ENCRYPTION_KEY`: Your encryption key. Generate a strong random one and store it in a safe place, as PG Back Web uses it to encrypt sensitive data.

- `PBW_POSTGRES_CONN_STRING`: The connection string for the PostgreSQL database that will store PG Back Web data.

- `PBW_LISTEN_HOST`: Optional. Host for the server to listen on, default 0.0.0.0

- `PBW_LISTEN_PORT`: Optional. Port for the server to listen on, default 8085

- `PBW_PATH_PREFIX`: Optional. Path prefix for the application URL. Use this when you want to serve the application under a subpath (e.g., `/pgbackweb`). Must start with `/` and not end with `/`. Default is empty.

- `TZ`: Optional. Your [timezone](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List). Default is `UTC`. This impacts logging, backup filenames and default timezone in the web interface.

## Screenshot

<img src="https://raw.githubusercontent.com/eduardolat/pgbackweb/main/assets/screenshot.png" />

## Key Features Explained

### Backup Management

- **Scheduled backups**: Configure backups with cron expressions for flexible scheduling (e.g., daily at 2 AM, weekly on Sundays)
- **Manual backups**: Trigger backups on-demand from the web interface
- **Backup duplication**: Clone existing backup configurations to quickly create similar backups
- **Backup activation**: Enable/disable backups without deleting them
- **Execution history**: View all backup executions with status, timestamps, file sizes, and download links

### Restoration

- **One-click restore**: Restore any backup to any configured database with a single click
- **Version-aware**: Automatically detects and uses the correct database version for restoration
- **Local and remote**: Restore from both local storage and S3-compatible storage
- **Restoration tracking**: Monitor restoration progress and view restoration history

### Webhooks

Configure webhooks to receive notifications for various events:

- **Database health events**: Get notified when databases become healthy or unhealthy
- **Destination health events**: Monitor storage destination availability
- **Execution events**: Receive notifications for successful or failed backup executions
- **Custom configuration**: Configure webhook URLs, HTTP methods (GET/POST), custom headers, and request bodies
- **Execution history**: View all webhook execution attempts with response details

### Health Checks

- **Automatic monitoring**: Regular health checks for all configured databases and destinations
- **Status tracking**: Visual indicators for healthy/unhealthy status
- **Test on demand**: Manually test database and destination connections
- **Bulk testing**: Test all databases or destinations at once

## Reset password

You can reset your PG Back Web password by running the following command in the server where PG Back Web is running:

### Docker

```bash
docker exec -it <container_name_or_id> sh -c change-password
```

You should replace `<container_name_or_id>` with the name or ID of the PG Back Web container, then just follow the instructions.

### Kubernetes

For Kubernetes deployments, use `kubectl exec` instead:

```bash
echo "user@example.com" | kubectl exec -i <pod_name> -n <namespace> -- change-password
```

Replace:

- `<pod_name>` with the name of your PG Back Web pod (e.g., `pgbackweb-5bc4c86566-ltdwq`)
- `<namespace>` with the namespace where PG Back Web is deployed (e.g., `database`)
- `user@example.com` with the email address of the user whose password you want to reset

The command will output a new randomly generated password that you can use to log in. You can change it after logging in through the web interface.

**Note:** The `-i` flag (without `-t`) is used because Kubernetes doesn't support interactive TTY in this context. The email is passed via stdin using `echo`.

## Next steps

In this link you can see a list of features that have been confirmed for future updates:

<a href="https://github.com/eduardolat/pgbackweb/issues?q=is%3Aissue+is%3Aopen+label%3A%22confirmed+next+step%22">
  Next steps ⏭️
</a>

## Sponsors

🙏 Thank you to the incredible sponsors for supporting this project! Your contributions help keep PG Back Web running and growing. If you'd like to join and become a sponsor, please visit the [sponsorship page](https://buymeacoffee.com/eduardolat) and be part of something great! 🚀

### 🥇 Gold Sponsors

<table>
  <tr>
    <td align="center">
      <a href="https://buymeacoffee.com/eduardolat">
        <img src="https://raw.githubusercontent.com/eduardolat/pgbackweb/refs/heads/develop/internal/view/static/images/plus-circle.png" height="150" alt="Become a gold sponsor"/>
        <br />
        Become a gold sponsor
      </a>
    </td>
  </tr>
</table>

### 🥈 Silver Sponsors

<table>
  <tr>
    <td align="center">
      <a href="https://fetchgoat.com?utm_source=pgbackweb&utm_medium=referral&utm_campaign=sponsorship">
        <img src="https://raw.githubusercontent.com/eduardolat/pgbackweb/refs/heads/develop/assets/sponsors/FetchGoat.png" height="100" alt="FetchGoat - Simplifying Logistics"/>
        <br />
        FetchGoat - Simplifying Logistics
      </a>
    </td>
    <td align="center">
      <a href="https://buymeacoffee.com/eduardolat">
        <img src="https://raw.githubusercontent.com/eduardolat/pgbackweb/refs/heads/develop/internal/view/static/images/plus-circle.png" height="100" alt="Become a silver sponsor"/>
        <br />
        Become a silver sponsor
      </a>
    </td>
  </tr>
</table>

### 🥉 Bronze Sponsors

<table>
  <tr>
    <td align="center">
      <a href="https://buymeacoffee.com/eduardolat">
        <img src="https://raw.githubusercontent.com/eduardolat/pgbackweb/refs/heads/develop/internal/view/static/images/plus-circle.png" height="80" alt="Become a bronze sponsor"/>
        <br />
        Become a bronze sponsor
      </a>
    </td>
  </tr>
</table>

## Join the Community

Got ideas to improve PG Back Web? Contribute to the project! Every suggestion and pull request is welcome.

## License

This project is 100% open source and is licensed under the AGPL v3 License - see the [LICENSE](LICENSE) file for details.

---

💖 **Love PG Back Web?** Give us a ⭐ on GitHub and share the project with your colleagues. Together, we can make PostgreSQL backups more accessible to everyone!
