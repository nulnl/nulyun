# Nul Yun

A modern, lightweight, self-hosted file management service with advanced features including TOTP two-factor authentication, file sharing, WebDAV support, and more.

![Version](https://img.shields.io/badge/version-0.0.0--beta6-blue)
![Go Version](https://img.shields.io/badge/go-1.25-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-Apache%202.0-green)

## Features

### Core Functionality
- 📁 **File Management**: Upload, download, rename, move, copy, and delete files and directories
- 🔍 **Advanced Search**: Fast file search with multiple criteria
- 📤 **Resumable Uploads**: TUS protocol support for reliable large file uploads
- 🖼️ **Media Preview**: Image thumbnails and previews with multiple quality levels
- 📊 **Disk Usage**: Real-time storage usage monitoring

### Security & Authentication
- 🔐 **Multiple Auth Methods**: JSON, Proxy, and NoAuth modes
-  **TOTP 2FA**: Time-based one-time password two-factor authentication
- 🎫 **JWT Tokens**: Secure session management with automatic renewal
- 🔒 **Permission System**: Granular file and administrative permissions

### Sharing & Collaboration
- 🔗 **Public Sharing**: Share files and folders with password protection
- ⏰ **Expiration Control**: Set share expiration dates
- 🌐 **WebDAV**: Mount as network drive on any platform
- 👥 **Multi-User**: Support for multiple users with isolated scopes

### Modern Architecture
- ⚡ **High Performance**: Built with Go for speed and efficiency
- 🎨 **Modern UI**: Vue.js SPA with responsive design
- 🐳 **Docker Ready**: Easy deployment with Docker and Docker Compose
- 📱 **Mobile Friendly**: Responsive design works on all devices
- 🌍 **Internationalization**: Multi-language support (i18n)

## Quick Start

### Using Binary

1. Download the latest release for your platform
2. Run with default settings:
```bash
./nulyun
```

3. Open browser and navigate to `http://localhost:8080`
4. Default credentials: `admin` / `admin` (change immediately!)

### Using Docker

```bash
docker run -d \
  --name nulyun \
  -p 8080:8080 \
  -v /path/to/data:/data \
  -v /path/to/config:/config \
  ghcr.io/nulnl/nulyun:latest
```

### Using Docker Compose

```yaml
version: '3.8'

services:
  nulyun:
    image: ghcr.io/nulnl/nulyun:latest
    container_name: nulyun
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
      - ./config:/config
      - ./database:/database
    environment:
      - TZ=UTC
    command:
      - --database=/database/nulyun.db
      - --root=/data
      - --address=0.0.0.0
      - --port=8080
```

## Building from Source

### Prerequisites

- Go 1.25 or higher
- Node.js 20 or higher
- pnpm (for frontend)

### Build Steps

```bash
# Clone the repository
git clone https://github.com/nulnl/nulyun.git
cd nulyun

# Build frontend and backend
make build

# Or build separately
make build-frontend  # Build Vue.js frontend
make build-backend   # Build Go backend

# Run tests
make test

# Binary will be in bin/nulyun
./bin/nulyun --help
```

## Configuration

### Command Line Flags

```bash
./nulyun \
  --config=/path/to/config.json \
  --database=/path/to/nulyun.db \
  --address=127.0.0.1 \
  --port=8080 \
  --root=/path/to/files \
  --baseURL=http://example.com \
  --cert=/path/to/cert.pem \
  --key=/path/to/key.pem \
  --log=stdout \
  --cacheDir=/path/to/cache \
  --tokenExpirationTime=2h \
  --disableThumbnails=false \
  --disablePreviewResize=false \
  --disableTOTP=false
```

### Configuration File

Create a `config.json` file:

```json
{
  "database": "./nulyun.db",
  "address": "0.0.0.0",
  "port": "8080",
  "root": "/data",
  "baseURL": "",
  "log": "stdout",
  "cacheDir": "/cache",
  "tokenExpirationTime": "2h",
  "totpTokenExpirationTime": "2m",
  "disableThumbnails": false,
  "disablePreviewResize": false,
  "disableTypeDetectionByHeader": false,
  "disableTOTP": false,
  "imageProcessors": 4,
  "username": "admin",
  "password": ""
}
```

Use with `./nulyun --config=config.json`

## Project Structure

Following Go standard project layout:

```
nulyun/
├── main.go                 # Application entry point
├── internal/               # Private application code
│   ├── auth/              # Authentication implementations
│   ├── errors/            # Custom error types
│   ├── files/             # File operations and services
│   ├── http/              # HTTP handlers and routing
│   ├── settings/          # Settings management
│   │   ├── global/        # Global server settings
│   │   ├── share/         # Share settings
│   │   ├── users/         # User settings
│   │   └── webdav/        # WebDAV settings
│   ├── storage/           # Storage interfaces
│   │   └── bolt/          # BoltDB implementation
│   └── version/           # Version information
├── www/                   # Frontend Vue.js application
│   ├── src/              # Vue.js source code
│   ├── public/           # Static assets
│   └── dist/             # Built frontend (embedded in binary)
├── docs/                  # Documentation
│   ├── API.md            # Complete API documentation
│   └── INTEGRATION.md    # Integration guide
├── Dockerfile            # Container build file
├── Makefile              # Build automation
└── go.mod                # Go dependencies
```

## API Documentation

Comprehensive API documentation is available for client developers:

- **[API Reference](docs/API.md)**: Complete REST API documentation with examples
- **[Integration Guide](docs/INTEGRATION.md)**: Quick integration guide for services and clients

### Key API Endpoints

- `POST /api/login` - Authentication
- `GET /api/resources{path}` - File listing and info
- `POST /api/resources{path}` - Upload files
- `GET /api/raw{path}` - Download files
- `GET /api/preview/{size}/{path}` - Image previews
- `GET /api/search` - Search files
- `POST /api/share` - Create shares
- `POST /api/tus` - TUS resumable uploads
- WebDAV endpoints for mounting as network drive

See [docs/API.md](docs/API.md) for complete documentation.

## Authentication Methods

### 1. JSON Authentication (Default)
Standard username/password stored in database.

### 2. Proxy Authentication
Use behind reverse proxy with `X-Forwarded-User` header.

### 3. NoAuth
Disable authentication (development only).

### 4. TOTP (Two-Factor)
Add additional security with time-based one-time passwords.

## WebDAV

Mount Nul Yun as a network drive:

### Windows
```
\\server-ip@8080\DavWWWRoot\
```

### macOS
```
Finder → Go → Connect to Server
http://server-ip:8080
```

### Linux
```bash
mount -t davfs http://server-ip:8080 /mnt/nulyun
```

## Development

### Running in Development

```bash
# Terminal 1: Run backend
go run main.go --database=./dev.db --root=./test-files

# Terminal 2: Run frontend dev server
cd www
pnpm install
pnpm run dev
```

Frontend will be available at `http://localhost:5173` with hot reload.

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
go test -cover ./...

# Run specific package tests
go test -v ./internal/files/...
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Vet code
make vet
```

## Deployment

### Production Checklist

- [ ] Use HTTPS with valid certificates
- [ ] Change default admin password
- [ ] Configure appropriate file permissions
- [ ] Set up regular database backups
- [ ] Configure log rotation
- [ ] Set up monitoring and health checks
- [ ] Review and adjust token expiration times
- [ ] Enable TOTP for admin accounts
- [ ] Configure appropriate disk quotas
- [ ] Set up firewall rules

### Reverse Proxy Example (Nginx)

```nginx
server {
    listen 443 ssl http2;
    server_name files.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    client_max_body_size 0;  # No upload size limit

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Timeouts for large uploads
        proxy_read_timeout 600s;
        proxy_send_timeout 600s;
    }
}
```

### Systemd Service

Create `/etc/systemd/system/nulyun.service`:

```ini
[Unit]
Description=Nul Yun File Server
After=network.target

[Service]
Type=simple
User=nulyun
Group=nulyun
WorkingDirectory=/opt/nulyun
ExecStart=/opt/nulyun/nulyun --config=/etc/nulyun/config.json
Restart=always
RestartSec=10

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/data /var/log/nulyun

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable nulyun
sudo systemctl start nulyun
sudo systemctl status nulyun
```

## Mobile Client Development

This project provides a complete REST API for building mobile clients. See [docs/API.md](docs/API.md) for Flutter examples and complete endpoint documentation.

### Flutter Quick Example

```dart
// See docs/API.md for complete Flutter client implementation
class NulYunClient {
  final String baseUrl;
  String? _token;
  
  Future<void> login(String username, String password) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/login'),
      body: jsonEncode({'username': username, 'password': password}),
    );
    _token = jsonDecode(response.body)['token'];
  }
}
```

## Troubleshooting

### Database Locked
If you see "database is locked" errors:
```bash
# Check if another process is using the database
lsof ./nulyun.db

# Or use a different database location
./nulyun --database=/tmp/nulyun.db
```

### Permission Issues
```bash
# Ensure proper ownership
chown -R nulyun:nulyun /data /database

# Set proper permissions
chmod 755 /data
chmod 600 /database/nulyun.db
```

### Frontend Not Loading
```bash
# Rebuild frontend
cd www && pnpm run build

# Check if dist folder exists
ls -la www/dist

# Rebuild binary with new frontend
make build
```

### Large File Upload Fails
- Check `client_max_body_size` in Nginx/Apache
- Increase timeouts in reverse proxy
- Use TUS protocol for files > 100MB
- Check disk space

## Performance Tuning

### Image Processing
```bash
# Increase image processors for faster thumbnail generation
./nulyun --imageProcessors=8
```

### File Caching
```bash
# Enable file cache for better performance
./nulyun --cacheDir=/var/cache/nulyun
```

### Database Optimization
- Regular VACUUM for BoltDB
- Keep database on SSD for better performance
- Use separate disk for data vs database

## Security Considerations

1. **Always use HTTPS in production**
2. **Change default credentials immediately**
3. **Enable 2FA for administrator accounts**
4. **Regular security updates**
5. **Implement rate limiting at reverse proxy level**
6. **Regular backup of database**
7. **Monitor failed login attempts**
8. **Use strong JWT signing keys**
9. **Implement file type restrictions if needed**
10. **Regular security audits**

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Format code (`make fmt`)
6. Run linter (`make lint`)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Write tests for new features
- Update documentation
- Use meaningful commit messages
- Keep PRs focused and small

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Go](https://golang.org/)
- Frontend powered by [Vue.js](https://vuejs.org/)
- UI framework: [Vuetify](https://vuetifyjs.com/)
- Database: [BoltDB](https://github.com/etcd-io/bbolt) via [Storm](https://github.com/asdine/storm)
- TUS Protocol: [tus.io](https://tus.io/)

## Roadmap

- [ ] S3 storage backend
- [ ] Elasticsearch integration for advanced search
- [ ] Collaborative editing
- [ ] Real-time file sync
- [ ] Video transcoding
- [ ] Audit logging
- [ ] API rate limiting
- [ ] Plugin system
- [ ] Mobile native apps (iOS/Android)

## Support

- 📖 Documentation: [docs/](docs/)
- 🐛 Issue Tracker: [GitHub Issues](https://github.com/nulnl/nulyun/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/nulnl/nulyun/discussions)

## Changelog

See [CHANGES.md](CHANGES.md) for version history and release notes.

---

**Made with ❤️ by the Nul team**
