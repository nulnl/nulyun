# WebDAV Feature Guide

## Overview

This project includes WebDAV support, allowing users to access the file system via the WebDAV protocol. Each user can create multiple WebDAV access tokens, and each token can be configured with a dedicated folder path and access permissions.

## Key Features

- ✅ Implements the standard WebDAV protocol using `golang.org/x/net/webdav`
- ✅ Each user can create multiple WebDAV tokens
- ✅ Each token can be configured with:
   - A dedicated folder path
   - Fine-grained permission controls (read/write/delete)
   - Status management (active/suspended)
- ✅ Full front-end management UI
- ✅ Command-line management tools

## API Endpoints

### WebDAV File Access

- **Path**: `/dav/`
- **Authentication methods**:
   - Bearer Token: `Authorization: Bearer <token>`
   - Basic Auth: username is the token, password may be empty
   - Query parameter: `?token=<token>`

### Token Management API

All API endpoints require user authentication.

- `GET /api/webdav/tokens` - List all tokens for the current user
- `GET /api/webdav/tokens/{id}` - Get details of a single token (returns the full token)
- `POST /api/webdav/tokens` - Create a new token
- `PUT /api/webdav/tokens/{id}` - Update a token
- `DELETE /api/webdav/tokens/{id}` - Delete a token
- `POST /api/webdav/tokens/{id}/suspend` - Suspend a token
- `POST /api/webdav/tokens/{id}/activate` - Activate a token

### Request Example

#### Create Token

```bash
curl -X POST http://localhost/api/webdav/tokens \
   -H "Authorization: Bearer <your-jwt-token>" \
   -H "Content-Type: application/json" \
   -d '{
      "name": "My WebDAV Access",
      "path": "/documents",
      "canRead": true,
      "canWrite": true,
      "canDelete": false
   }'
```

## Web UI Usage

1. Sign in to the nul yun
2. Go to **Settings** -> **WebDAV**
3. Click the **Create Token** button
4. Fill in the form:
    - **Name**: A descriptive name for the token
    - **Path**: The folder path the token is restricted to (defaults to `/`)
    - **Permissions**: Check required permissions (read/write/delete)
5. Click **Create**
6. **Important**: Save the full token shown after creation — it will only be displayed once!

### Token Management Actions

- **View details**: Click the info icon to view token details and the WebDAV URL
- **Edit**: Modify the token's name, path, and permissions
- **Suspend/Activate**: Temporarily disable or re-enable a token
- **Delete**: Permanently remove a token

## CLI Usage

### List a user's tokens

```bash
nulyun webdav ls <username>
```

### Create a token

```bash
nulyun webdav add <username> <name> <path> [--read] [--write] [--delete]
```

Example:
```bash
nulyun webdav add admin "MyWebDAV" "/documents" --read --write
```

### Delete a token

```bash
nulyun webdav rm <username> <token-id>
```

### Suspend a token

```bash
nulyun webdav suspend <username> <token-id>
```

### Activate a token

```bash
nulyun webdav activate <username> <token-id>
```

## WebDAV Client Configuration

### Windows Explorer

1. Open "This PC"
2. Right-click and choose "Add a network location"
3. Enter the WebDAV URL: `http://your-server/dav/`
4. Authentication:
   - Username: your token string
   - Password: leave empty or provide any character

### macOS Finder

1. Finder -> Go -> Connect to Server
2. Enter: `http://your-server/dav/`
3. In the authentication dialog:
   - Username: your token
   - Password: leave empty or provide any character

### Linux (davfs2)

```bash
# Install davfs2
sudo apt-get install davfs2  # Ubuntu/Debian
sudo yum install davfs2       # CentOS/RHEL

# Mount
sudo mount -t davfs http://your-server/dav/ /mnt/webdav

# When prompted for credentials:
# Username: <your-token>
# Password: (leave empty)
```

### Third-party clients

- **WinSCP** (Windows)
- **Cyberduck** (macOS/Windows)
- **Total Commander** (Windows)
- **Transmit** (macOS)

Configuration:
- Protocol: WebDAV (HTTP/HTTPS)
- Host: `your-server`
- Path: `/dav/`
- Username: WebDAV token
- Password: leave empty

## Permissions

### Read permission (canRead)

Allowed operations:
- `GET` - Download files
- `HEAD` - Get file info
- `OPTIONS` - Get server options
- `PROPFIND` - List directory and file properties

### Write permission (canWrite)

Allowed operations:
- `PUT` - Upload files
- `POST` - Create resources
- `PATCH` - Partial updates
- `MKCOL` - Create directories
- `COPY` - Copy files/directories
- `MOVE` - Move/rename files/directories

### Delete permission (canDelete)

Allowed operations:
- `DELETE` - Remove files or directories

## Security Recommendations

1. **Token security**
   - Tokens are shown only once after creation; store them securely
   - Rotate tokens periodically
   - Remove tokens when not in use

2. **Least privilege**
   - Grant only necessary permissions
   - Create different tokens for different purposes
   - Restrict token access paths

3. **Use HTTPS**
   - Enable HTTPS in production
   - Avoid transmitting tokens over insecure networks

4. **Monitoring and auditing**
   - Regularly review active tokens
   - Monitor for abnormal access patterns
   - Suspend or delete suspicious tokens promptly

## Troubleshooting

### Connection refused

- Check whether the token is active
- Ensure the token has not expired or been deleted
- Verify network connectivity and firewall settings

### Permission errors

- Check whether the token has the required permissions
- Ensure the accessed path is within the token's allowed scope
- Verify the user's base file system permissions

### Cannot list files

- Ensure the token has read permission
- Check that the path is correct
- Verify that the folder exists

## Implementation Details

### Backend layout

```
webdav/
├── token.go      # Token data model
├── storage.go    # Token storage interface
└── webdav.go     # WebDAV protocol handler
storage/bolt/
└── webdav.go     # Bolt DB implementation
http/
└── webdav.go     # HTTP API handlers
```

### Frontend layout

```
www/src/
├── api/
│   └── webdav.ts           # API client
└── views/settings/
    └── WebDAV.vue          # Token management UI
```

## Database Structure

Token storage structure in Bolt DB:

```go
type Token struct {
      ID        uint        // auto-increment primary key
      UserID    uint        // owning user ID (indexed)
      Name      string      // token name
      Token     string      // access token (unique)
      Path      string      // restricted path
      CanRead   bool        // read permission
      CanWrite  bool        // write permission
      CanDelete bool        // delete permission
      Status    TokenStatus // status (active/suspended)
      CreatedAt time.Time   // creation time
      UpdatedAt time.Time   // update time
}
```

## FAQ

**Q: Can multiple tokens be created for the same path?**  
A: Yes. A user may create multiple tokens for the same path with different permissions.

**Q: Do tokens expire?**  
A: Tokens do not expire automatically in the current implementation, but they can be suspended or deleted manually.

**Q: What is the difference between suspend and delete?**  
A: Suspend temporarily disables a token and it can be reactivated; delete permanently removes it and requires recreation.

**Q: Can I see token usage history?**  
A: Usage logging is not included in the current version; it can be added in a future release.

## Contributing

Contributions, issues and suggestions are welcome!
