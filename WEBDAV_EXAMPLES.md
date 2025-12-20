# WebDAV Usage Examples

## Quick Start

### 1. Create a WebDAV token (Web UI)

1. Sign in to nulyun
2. Click the settings button in the top-right
3. Select "WebDAV" from the sidebar
4. Click the "Create Token" button
5. Fill in the form:
   - **Name**: "My Documents Access"
   - **Path**: "/documents"
   - **Permissions**: check "Read" and "Write"
6. Click "Create"
7. Important: copy and store the full token — it will only be shown once.

### 2. Create a token from the CLI

```bash
# Create a read/write token for user `admin` that accesses /documents
nulyun webdav add admin "My Documents" "/documents" --read --write
```

### Mounting examples (macOS / Windows / Linux)

Use the WebDAV URL: `http://your-server.com:8080/dav/`

Common CLI examples:

```bash
nulyun webdav ls admin
nulyun webdav add admin "Read Only" "/" --read --no-write --no-delete
nulyun webdav rm admin 1
nulyun webdav suspend admin 1
nulyun webdav activate admin 1
```

### cURL examples

```bash
curl -X PROPFIND \
  -H "Authorization: Bearer <your-token>" \
  http://your-server.com:8080/dav/

# Download a file
curl -X GET \
  -H "Authorization: Bearer <your-token>" \
  http://your-server.com:8080/dav/file.txt

# Upload a file
curl -X PUT \
  -H "Authorization: Bearer <your-token>" \
  --data-binary @local-file.txt \
  http://your-server.com:8080/dav/remote-file.txt
```

### Programming example (Node.js)

```javascript
const { createClient } = require("webdav");

const client = createClient(
  "http://your-server.com:8080/dav/",
  {
    username: "<your-token>",
    password: ""
  }
);

// List files
async function listFiles() {
  const contents = await client.getDirectoryContents("/");
  console.log(contents);
}
```

## CLI Reference

- List tokens: `nulyun webdav ls <username>`
- Add token: `nulyun webdav add <username> <name> <path> [--read] [--write] [--delete]`
- Remove token: `nulyun webdav rm <username> <token-id>`
- Suspend token: `nulyun webdav suspend <username> <token-id>`
- Activate token: `nulyun webdav activate <username> <token-id>`

## Notes

Use HTTPS in production, store tokens securely, and follow the principle of least privilege.

---

## Additional Mounting Examples and Tools

### Windows (map network drive)

1. Open "This PC"
2. Right-click and choose "Add a network location"
3. Click "Next" -> "Choose a custom network location"
4. Enter the server address: `http://your-server.com:8080/dav/`
5. In the authentication dialog:
   - Username: paste your token (for example `dGVzdC10b2tlbi1hYmMxMjM0NTY3ODk=`)
   - Password: leave empty or enter any value
6. Name the network location: `Nul Yun WebDAV`
7. Done — you can now use it like a local folder.

### macOS (Finder)

1. Open Finder and press Command+K
2. Enter the server address: `http://your-server.com:8080/dav/`
3. Alternatively, open the URL using the token in the userinfo: `open "http://<your-token>@your-server.com:8080/dav/"`

When prompted for credentials:
- Username: your token
- Password: leave empty or enter any value

### macOS (CLI mount)

```bash
# Create mount point
mkdir -p ~/WebDAV

# Mount via mount_webdav
mount_webdav -i http://your-server.com:8080/dav/ ~/WebDAV

# When prompted:
# Username: <your-token>
# Password: (press Enter to skip)
```

### Linux (davfs2)

```bash
sudo apt-get install davfs2  # Ubuntu/Debian
sudo yum install davfs2      # CentOS/RHEL

# Create mount point
sudo mkdir -p /mnt/webdav

# Mount
sudo mount -t davfs http://your-server.com:8080/dav/ /mnt/webdav

# Example fstab entry to mount at boot:
echo "http://your-server.com:8080/dav/ /mnt/webdav davfs user,noauto 0 0" | sudo tee -a /etc/fstab
# Add token to davfs2 secrets
echo "<your-token>" | sudo tee -a /etc/davfs2/secrets
```

### Client tools

#### Cyberduck (Windows/Mac)

1. Click "Open Connection"
2. Protocol: WebDAV (HTTP)
3. Server: `your-server.com`
4. Port: `8080`
5. Path: `/dav/`
6. Username: your token
7. Password: leave empty

#### WinSCP (Windows)

1. Open WinSCP
2. File protocol: WebDAV
3. Host name: `your-server.com`
4. Port number: `8080`
5. Username: your token
6. Password: leave empty

### Programmatic access

#### cURL

```bash
curl -X PROPFIND \
  -H "Authorization: Bearer <your-token>" \
  http://your-server.com:8080/dav/

# Download file
curl -X GET \
  -H "Authorization: Bearer <your-token>" \
  http://your-server.com:8080/dav/file.txt

# Upload file
curl -X PUT \
  -H "Authorization: Bearer <your-token>" \
  --data-binary @local-file.txt \
  http://your-server.com:8080/dav/remote-file.txt

# Delete file
curl -X DELETE \
  -H "Authorization: Bearer <your-token>" \
  http://your-server.com:8080/dav/file.txt
```

#### Python (webdav3)

```python
from webdav3.client import Client

options = {
    'webdav_hostname': "http://your-server.com:8080/dav/",
    'webdav_login':    "<your-token>",
    'webdav_password': ""
}

client = Client(options)

# List files
files = client.list()
print(files)

# Download
client.download_sync(remote_path="file.txt", local_path="local-file.txt")

# Upload
client.upload_sync(remote_path="remote-file.txt", local_path="local-file.txt")

# Delete
client.clean("file.txt")
```

#### Node.js

```javascript
const { createClient } = require("webdav");

const client = createClient(
  "http://your-server.com:8080/dav/",
  {
    username: "<your-token>",
    password: ""
  }
);

// List files
async function listFiles() {
  const contents = await client.getDirectoryContents("/");
  console.log(contents);
}

// Download file
async function downloadFile() {
  const buffer = await client.getFileContents("/file.txt");
  require("fs").writeFileSync("local-file.txt", buffer);
}

// Upload file
async function uploadFile() {
  const data = require("fs").readFileSync("local-file.txt");
  await client.putFileContents("/remote-file.txt", data);
}

// Delete file
async function deleteFile() {
  await client.deleteFile("/file.txt");
}
```

### 8. Token Management

#### List all tokens

```bash
# CLI
nulyun webdav ls admin

# Output:
# ID  Name         Token              Path        Read  Write  Delete  Status   Created At
# 1   My Documents dGVzdC10b2tlbi...  /documents  true  true  false   active   2024-01-15 10:30:00
```

#### Suspend a token

```bash
# Temporarily disable a token
nulyun webdav suspend admin 1

# Output:
# WebDAV token (ID: 1) suspended
```

#### Reactivate a token

```bash
# Reactivate a token
nulyun webdav activate admin 1

# Output:
# WebDAV token (ID: 1) activated
```

#### Remove a token

```bash
# Permanently remove a token
nulyun webdav rm admin 1

# Output:
# WebDAV token (ID: 1) deleted
```

### 9. Security Best Practices

#### Create separate tokens for different purposes

```bash
# Read-only token
nulyun webdav add admin "Read Only Access" "/" --read --no-write --no-delete

# Token limited to photos directory
nulyun webdav add admin "Photo Uploads" "/photos" --read --write --no-delete

# Temporary access token (delete after use)
nulyun webdav add admin "Temporary Share" "/shared" --read --no-write --no-delete
```

#### Review tokens regularly

```bash
# List all tokens
nulyun webdav ls admin

# Remove tokens that are no longer used
nulyun webdav rm admin 5
nulyun webdav rm admin 6
```

### 10. Troubleshooting

#### Connection failures

```bash
# Test connection
curl -v -X PROPFIND \
  -H "Authorization: Bearer <your-token>" \
  http://your-server.com:8080/dav/

# Check whether the token is active
nulyun webdav ls admin | grep <token-id>
```

#### Permission errors

Check token permissions:
```bash
nulyun webdav ls admin
# Look at the "Read", "Write", "Delete" columns in the output
```

If you need to update permissions, edit the token in the web UI.

#### Path issues

Ensure the requested path is within the token's allowed scope:
```bash
# If a token's path is "/documents"
# then it can only access /documents and its subdirectories

# Correct:
http://your-server.com:8080/dav/report.pdf

# Incorrect (if report.pdf is not under /documents):
http://your-server.com:8080/dav/../other/report.pdf
```

## Advanced Use Cases

### Use Case 1: Team document sharing

Create read-only access for team members:
```bash
nulyun webdav add team_member "Team Docs Read-Only" "/team-docs" --read --no-write --no-delete
```

### Use Case 2: Automated backups

Use rsync and davfs2 for scheduled backups:
```bash
#!/bin/bash
# backup.sh

# Mount WebDAV
mount -t davfs http://server/dav/ /mnt/webdav

# Sync files
rsync -av /local/data/ /mnt/webdav/backup/

# Unmount
umount /mnt/webdav
```

### Use Case 3: Automatic photo uploads

Use inotify to watch a directory and upload new photos:
```bash
#!/bin/bash
# watch_upload.sh

inotifywait -m /path/to/photos -e create -e moved_to |
while read dir action file; do
  curl -X PUT \
    -H "Authorization: Bearer <token>" \
    --data-binary @"$dir$file" \
    "http://server/dav/photos/$file"
done
```

## Summary

WebDAV provides flexible file access, and token-based management enables fine-grained permission control. Choose appropriate access methods and permissions for your needs to keep data secure while maintaining convenient access.
