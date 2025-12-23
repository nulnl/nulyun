# Nul Yun API Documentation

**Version**: v0.0.0-beta6  
**Base URL**: `http://your-server:8080` (or your configured base URL)  
**API Prefix**: `/api`

This document provides complete API reference for building client applications (web, mobile, desktop) that integrate with Nul Yun file service.

---

## Table of Contents

1. [Authentication](#authentication)
2. [User Management](#user-management)
3. [File Operations](#file-operations)
4. [File Upload (TUS Protocol)](#file-upload-tus-protocol)
5. [Preview & Raw File Access](#preview--raw-file-access)
6. [Search](#search)
7. [Sharing](#sharing)
8. [Public Access](#public-access)
9. [Settings](#settings)
10. [WebDAV](#webdav)
11. [Passkey (WebAuthn)](#passkey-webauthn)
12. [TOTP (Two-Factor Authentication)](#totp-two-factor-authentication)
13. [Error Handling](#error-handling)
14. [Flutter Client Examples](#flutter-client-examples)

---

## Authentication

### Login

Authenticate and receive a JWT token.

**Endpoint**: `POST /api/login`

**Request Body**:
```json
{
  "username": "admin",
  "password": "secret"
}
```

**Response** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "otp": false
}
```

**Response with TOTP** (if two-factor is enabled):
```json
{
  "token": "temporary_token",
  "otp": true
}
```

If `otp: true`, continue to [TOTP verification](#verify-totp).

**Error Responses**:
- `401 Unauthorized`: Invalid credentials
- `403 Forbidden`: Account disabled or permission denied
- `500 Internal Server Error`: Server error

**Flutter Example**:
```dart
Future<String> login(String username, String password) async {
  final response = await http.post(
    Uri.parse('$baseUrl/api/login'),
    headers: {'Content-Type': 'application/json'},
    body: jsonEncode({'username': username, 'password': password}),
  );
  
  if (response.statusCode == 200) {
    final data = jsonDecode(response.body);
    if (data['otp'] == true) {
      // Handle TOTP flow
      return data['token']; // Temporary token
    }
    return data['token'];
  }
  throw Exception('Login failed');
}
```

---

### Using the Token

Include the JWT token in all authenticated requests using the `X-Auth` header:

```
X-Auth: <your-jwt-token>
```

Alternatively, for GET requests, you can use a cookie named `auth`.

**Token Expiration**: Default is 2 hours. Check response headers:
- `X-Renew-Token: true` — Token expires soon or user data changed, renew it

---

### Renew Token

Refresh your JWT token before expiration.

**Endpoint**: `POST /api/renew`

**Headers**:
```
X-Auth: <current-token>
```

**Response** (200 OK):
```json
{
  "token": "new_jwt_token",
  "otp": false
}
```

---

### Signup

Create a new user account (if signup is enabled by server settings).

**Endpoint**: `POST /api/signup`

**Request Body**:
```json
{
  "username": "newuser",
  "password": "securepassword"
}
```

**Response**:
- `200 OK`: Account created successfully
- `400 Bad Request`: Invalid username/password
- `405 Method Not Allowed`: Signup disabled
- `409 Conflict`: Username already exists

---

## User Management

All user management endpoints require authentication with admin privileges.

### List Users

**Endpoint**: `GET /api/users`

**Headers**: `X-Auth: <admin-token>`

**Response** (200 OK):
```json
[
  {
    "id": 1,
    "username": "admin",
    "scope": "/path/to/user/files",
    "locale": "en",
    "viewMode": "list",
    "singleClick": false,
    "perm": {
      "admin": true,
      "execute": true,
      "create": true,
      "rename": true,
      "modify": true,
      "delete": true,
      "share": true,
      "download": true
    },
    "lockPassword": false,
    "hideDotfiles": false,
    "dateFormat": false,
    "otpEnabled": false,
    "passkeyEnabled": false
  }
]
```

---

### Get User

**Endpoint**: `GET /api/users/{id}`

**Headers**: `X-Auth: <admin-token>`

**Response** (200 OK): Same structure as user object above.

---

### Create User

**Endpoint**: `POST /api/users`

**Headers**: 
```
X-Auth: <admin-token>
Content-Type: application/json
```

**Request Body**:
```json
{
  "username": "newuser",
  "password": "password123",
  "scope": "/",
  "locale": "en",
  "viewMode": "list",
  "perm": {
    "admin": false,
    "execute": true,
    "create": true,
    "rename": true,
    "modify": true,
    "delete": true,
    "share": true,
    "download": true
  },
  "lockPassword": false,
  "hideDotfiles": false
}
```

**Response**: `200 OK` with created user object.

---

### Update User

**Endpoint**: `PUT /api/users/{id}`

**Headers**: 
```
X-Auth: <admin-token>
Content-Type: application/json
```

**Request Body**: Same as Create User (all fields optional for partial update).

**Response**: `200 OK` with updated user object.

---

### Delete User

**Endpoint**: `DELETE /api/users/{id}`

**Headers**: `X-Auth: <admin-token>`

**Response**: `200 OK`

---

## File Operations

### List Files / Get File Info

**Endpoint**: `GET /api/resources{path}`

**Path Parameter**: URL-encoded file/directory path, e.g., `/api/resources/Documents/file.txt`

**Headers**: `X-Auth: <token>`

**Query Parameters** (for directories):
- `sort`: Sort field (name, size, modified)
- `order`: asc or desc
- `limit`: Number of items per page
- `offset`: Pagination offset

**Response for File** (200 OK):
```json
{
  "name": "file.txt",
  "size": 1024,
  "modified": "2025-12-23T10:00:00Z",
  "mode": 33188,
  "isDir": false,
  "isSymlink": false,
  "type": "text/plain",
  "path": "/Documents/file.txt"
}
```

**Response for Directory** (200 OK):
```json
{
  "name": "Documents",
  "size": 4096,
  "modified": "2025-12-23T10:00:00Z",
  "mode": 16877,
  "isDir": true,
  "path": "/Documents",
  "items": [
    {
      "name": "file1.txt",
      "size": 1024,
      "modified": "2025-12-23T09:00:00Z",
      "isDir": false,
      "type": "text/plain"
    }
  ],
  "numDirs": 5,
  "numFiles": 10
}
```

---

### Upload File (Simple)

For small files, use multipart form upload.

**Endpoint**: `POST /api/resources{path}`

**Path Parameter**: Destination directory path

**Headers**: 
```
X-Auth: <token>
Content-Type: multipart/form-data
```

**Form Data**:
- `file`: The file to upload (binary)
- `override`: "true" to overwrite existing file (optional)

**Response**: `200 OK`

**Flutter Example**:
```dart
Future<void> uploadFile(String filePath, String destinationPath) async {
  var request = http.MultipartRequest(
    'POST',
    Uri.parse('$baseUrl/api/resources$destinationPath'),
  );
  request.headers['X-Auth'] = token;
  request.files.add(await http.MultipartFile.fromPath('file', filePath));
  
  var response = await request.send();
  if (response.statusCode != 200) {
    throw Exception('Upload failed');
  }
}
```

---

### Create Directory

**Endpoint**: `POST /api/resources{path}`

**Path Parameter**: New directory path

**Headers**: 
```
X-Auth: <token>
Content-Type: application/json
```

**Request Body**:
```json
{
  "action": "mkdir"
}
```

**Response**: `200 OK`

---

### Delete File/Directory

**Endpoint**: `DELETE /api/resources{path}`

**Path Parameter**: File or directory path to delete

**Headers**: `X-Auth: <token>`

**Response**: `200 OK`

---

### Rename/Move File

**Endpoint**: `PATCH /api/resources{path}`

**Path Parameter**: Current file path

**Headers**: 
```
X-Auth: <token>
Content-Type: application/json
```

**Request Body**:
```json
{
  "action": "rename",
  "destination": "/new/path/newname.txt"
}
```

**Actions**:
- `rename`: Move/rename file
- `copy`: Copy file to destination
- `delete`: Delete (alternative to DELETE method)

**Response**: `200 OK`

---

### Copy File

**Endpoint**: `PATCH /api/resources{path}`

**Request Body**:
```json
{
  "action": "copy",
  "destination": "/destination/path/filename.txt"
}
```

---

### Bulk Operations

**Endpoint**: `PATCH /api/resources`

**Request Body**:
```json
{
  "action": "delete",
  "files": [
    "/path/to/file1.txt",
    "/path/to/file2.txt"
  ]
}
```

**Supported Actions**: `delete`, `copy`, `rename`

---

## File Upload (TUS Protocol)

For resumable uploads of large files, use the TUS protocol.

### Create Upload

**Endpoint**: `POST /api/tus`

**Headers**:
```
X-Auth: <token>
Tus-Resumable: 1.0.0
Upload-Length: 1048576
Upload-Metadata: filename dGVzdC5tcDQ=
```

`Upload-Metadata` is base64-encoded key-value pairs.

**Response** (201 Created):
```
Location: /api/tus/upload-id
Tus-Resumable: 1.0.0
```

---

### Upload Chunk

**Endpoint**: `PATCH /api/tus/{upload-id}`

**Headers**:
```
X-Auth: <token>
Tus-Resumable: 1.0.0
Upload-Offset: 0
Content-Type: application/offset+octet-stream
```

**Body**: Binary chunk data

**Response** (204 No Content):
```
Upload-Offset: 524288
```

---

### Check Upload Status

**Endpoint**: `HEAD /api/tus/{upload-id}`

**Headers**:
```
X-Auth: <token>
Tus-Resumable: 1.0.0
```

**Response**:
```
Upload-Offset: 524288
Upload-Length: 1048576
```

---

### Delete Upload

**Endpoint**: `DELETE /api/tus/{upload-id}`

**Headers**: `X-Auth: <token>`

**Response**: `204 No Content`

---

## Preview & Raw File Access

### Get Raw File

Download file content directly.

**Endpoint**: `GET /api/raw{path}`

**Path Parameter**: File path

**Headers**: 
```
X-Auth: <token>
```

**Query Parameters**:
- `inline`: Set to "true" to display in browser instead of download
- `checksum`: Algorithm (md5, sha1, sha256, sha512) to return file hash

**Response**: Binary file content with appropriate `Content-Type` and `Content-Disposition` headers.

**Response Headers**:
```
Content-Type: application/octet-stream
Content-Disposition: attachment; filename="file.txt"
X-Content-SHA256: <hash> (if checksum requested)
```

---

### Get Image Preview/Thumbnail

**Endpoint**: `GET /api/preview/{size}/{path}`

**Path Parameters**:
- `size`: `thumb` (thumbnail), `big` (large preview), or `original`
- `path`: Image file path

**Headers**: `X-Auth: <token>`

**Query Parameters**:
- `quality`: Image quality (high, medium, low)

**Response**: Image binary data (JPEG/PNG)

**Example**: `GET /api/preview/thumb/Photos/vacation.jpg?quality=medium`

---

### Get Subtitle

For video files with subtitle tracks.

**Endpoint**: `GET /api/subtitle{path}`

**Path Parameter**: Video file path

**Headers**: `X-Auth: <token>`

**Query Parameters**:
- `index`: Subtitle track index (default: 0)

**Response**: Subtitle file content (SRT, VTT, etc.)

---

## Search

**Endpoint**: `GET /api/search`

**Headers**: `X-Auth: <token>`

**Query Parameters**:
- `query`: Search term (filename, extension, content)
- `path`: Base directory to search in (default: user root)
- `sort`: Sort field
- `order`: asc/desc

**Response** (200 OK):
```json
{
  "items": [
    {
      "name": "document.pdf",
      "path": "/Documents/document.pdf",
      "size": 204800,
      "modified": "2025-12-20T15:30:00Z",
      "isDir": false,
      "type": "application/pdf"
    }
  ],
  "numDirs": 0,
  "numFiles": 1
}
```

---

## Sharing

### List Shares

**Endpoint**: `GET /api/shares`

**Headers**: `X-Auth: <token>`

**Response** (200 OK):
```json
[
  {
    "hash": "abc123xyz",
    "path": "/Documents/report.pdf",
    "userID": 1,
    "expire": 1735689600,
    "password": false
  }
]
```

---

### Get Share Details

**Endpoint**: `GET /api/share/{hash}`

**Headers**: `X-Auth: <token>`

**Response** (200 OK):
```json
{
  "hash": "abc123xyz",
  "path": "/Documents/report.pdf",
  "userID": 1,
  "expire": 1735689600,
  "password": false
}
```

---

### Create Share

**Endpoint**: `POST /api/share`

**Headers**: 
```
X-Auth: <token>
Content-Type: application/json
```

**Request Body**:
```json
{
  "path": "/Documents/report.pdf",
  "password": "",
  "expires": "2025-12-31T23:59:59Z",
  "unit": "hours",
  "number": 24
}
```

**Fields**:
- `path`: File/directory to share (required)
- `password`: Optional password protection
- `expires`: Expiration date (ISO 8601)
- `unit`: hours, days, months
- `number`: Number of units

**Response** (200 OK):
```json
{
  "hash": "newshare123",
  "url": "http://server:8080/share/newshare123"
}
```

---

### Delete Share

**Endpoint**: `DELETE /api/share/{hash}`

**Headers**: `X-Auth: <token>`

**Response**: `200 OK`

---

## Public Access

Public share endpoints (no authentication required or uses share-specific auth).

### Access Public Share

**Endpoint**: `GET /api/public/share/{hash}{path}`

**Path Parameters**:
- `hash`: Share hash
- `path`: Optional subpath within shared directory

**Headers** (if password-protected):
```
X-Share-Password: <password>
```

**Response**: File info or directory listing (same format as `/api/resources`).

---

### Download from Public Share

**Endpoint**: `GET /api/public/dl/{hash}{path}`

**Path Parameters**:
- `hash`: Share hash
- `path`: File path within share

**Headers** (if password-protected):
```
X-Share-Password: <password>
```

**Query Parameters**:
- `inline`: "true" to display instead of download
- `token`: Share token (alternative to header)

**Response**: File binary content

---

## Settings

### Get Server Settings

**Endpoint**: `GET /api/settings`

**Headers**: `X-Auth: <admin-token>`

**Response** (200 OK):
```json
{
  "signup": false,
  "createUserDir": false,
  "userHomeBasePath": "/users",
  "defaults": {
    "scope": "/",
    "locale": "en",
    "viewMode": "list",
    "singleClick": false,
    "perm": {
      "admin": false,
      "execute": true,
      "create": true,
      "rename": true,
      "modify": true,
      "delete": true,
      "share": true,
      "download": true
    }
  },
  "branding": {
    "name": "Nul Yun",
    "disableExternal": false,
    "files": "/",
    "theme": {
      "primary": "#2979ff",
      "secondary": "#0066cc"
    }
  },
  "tus": {
    "chunkSize": 10485760
  },
  "commands": [],
  "shell": []
}
```

---

### Update Settings

**Endpoint**: `PUT /api/settings`

**Headers**: 
```
X-Auth: <admin-token>
Content-Type: application/json
```

**Request Body**: Same structure as Get Settings response (partial updates supported).

**Response**: `200 OK`

---

## WebDAV

WebDAV endpoints are mounted at the root level (not under `/api`).

### WebDAV Root

**Endpoint**: `PROPFIND /`

**Headers**:
```
Authorization: Basic <base64(username:password)>
Depth: 0
```

Or use WebDAV token authentication:
```
Authorization: Bearer <webdav-token>
```

Standard WebDAV methods are supported:
- `PROPFIND`: List files/properties
- `GET`: Download file
- `PUT`: Upload file
- `DELETE`: Delete file
- `MKCOL`: Create directory
- `COPY`: Copy file
- `MOVE`: Move/rename file

### Create WebDAV Token

**Endpoint**: `POST /api/webdav/token`

**Headers**: `X-Auth: <token>`

**Response** (200 OK):
```json
{
  "token": "webdav-token-xyz"
}
```

---

## Passkey (WebAuthn)

### List Passkeys

**Endpoint**: `GET /api/passkeys`

**Headers**: `X-Auth: <token>`

**Response** (200 OK):
```json
[
  {
    "id": 1,
    "name": "iPhone 15",
    "credentialID": "base64-encoded-credential-id",
    "created": "2025-12-01T10:00:00Z"
  }
]
```

---

### Begin Passkey Registration

**Endpoint**: `POST /api/passkeys/register/begin`

**Headers**: 
```
X-Auth: <token>
Content-Type: application/json
```

**Request Body**:
```json
{
  "name": "My Security Key"
}
```

**Response** (200 OK):
```json
{
  "publicKey": {
    "challenge": "base64-challenge",
    "rp": {
      "name": "Nul Yun",
      "id": "localhost"
    },
    "user": {
      "id": "base64-user-id",
      "name": "username",
      "displayName": "User Display Name"
    },
    "pubKeyCredParams": [...],
    "timeout": 60000,
    "authenticatorSelection": {...}
  }
}
```

Use this data with WebAuthn JavaScript API:
```javascript
const credential = await navigator.credentials.create({
  publicKey: response.publicKey
});
```

---

### Finish Passkey Registration

**Endpoint**: `POST /api/passkeys/register/finish`

**Headers**: 
```
X-Auth: <token>
Content-Type: application/json
```

**Request Body**: WebAuthn credential response from browser

**Response**: `200 OK` with saved passkey details

---

### Delete Passkey

**Endpoint**: `DELETE /api/passkeys/{id}`

**Headers**: `X-Auth: <token>`

**Response**: `200 OK`

---

### Passkey Login - Begin

**Endpoint**: `POST /api/passkey/login/begin`

**Headers**: `Content-Type: application/json`

**Request Body**:
```json
{
  "username": "admin"
}
```

**Response** (200 OK):
```json
{
  "publicKey": {
    "challenge": "base64-challenge",
    "timeout": 60000,
    "rpId": "localhost",
    "allowCredentials": [...]
  }
}
```

---

### Passkey Login - Finish

**Endpoint**: `POST /api/passkey/login/finish`

**Headers**: `Content-Type: application/json`

**Request Body**: WebAuthn assertion response

**Response** (200 OK):
```json
{
  "token": "jwt-token",
  "otp": false
}
```

---

## TOTP (Two-Factor Authentication)

### Enable TOTP

**Endpoint**: `POST /api/users/{id}/otp`

**Headers**: `X-Auth: <admin-token>`

**Response** (200 OK):
```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qrcode": "data:image/png;base64,iVBORw0KGgo...",
  "recoveryCodes": [
    "12345678",
    "87654321"
  ]
}
```

Display the QR code for the user to scan with their authenticator app.

---

### Verify TOTP (Complete Setup)

**Endpoint**: `POST /api/users/{id}/otp/check`

**Headers**: 
```
X-Auth: <admin-token>
Content-Type: application/json
```

**Request Body**:
```json
{
  "code": "123456"
}
```

**Response**: `200 OK` if code is valid

---

### Verify TOTP (Login)

Used when `otp: true` is returned from login.

**Endpoint**: `POST /api/login/otp`

**Headers**: 
```
X-Auth: <temporary-token>
Content-Type: application/json
```

**Request Body**:
```json
{
  "code": "123456"
}
```

Or use recovery code:
```json
{
  "recoveryCode": "12345678"
}
```

**Response** (200 OK):
```json
{
  "token": "full-access-jwt-token",
  "otp": false
}
```

---

### Get TOTP Status

**Endpoint**: `GET /api/users/{id}/otp`

**Headers**: `X-Auth: <admin-token>`

**Response** (200 OK):
```json
{
  "enabled": true,
  "verified": true
}
```

---

### Disable TOTP

**Endpoint**: `DELETE /api/users/{id}/otp`

**Headers**: `X-Auth: <admin-token>`

**Response**: `200 OK`

---

### Reset TOTP

Generate new secret and recovery codes.

**Endpoint**: `POST /api/users/{id}/otp/reset`

**Headers**: `X-Auth: <admin-token>`

**Response**: Same as Enable TOTP

---

### Regenerate Recovery Codes

**Endpoint**: `POST /api/users/{id}/otp/recovery`

**Headers**: `X-Auth: <admin-token>`

**Response** (200 OK):
```json
{
  "recoveryCodes": [
    "11111111",
    "22222222"
  ]
}
```

---

## Error Handling

All errors follow this format:

**Response** (4xx or 5xx):
```json
{
  "error": "Error message description"
}
```

### Common HTTP Status Codes

- `200 OK`: Success
- `201 Created`: Resource created
- `204 No Content`: Success, no response body
- `400 Bad Request`: Invalid request format or parameters
- `401 Unauthorized`: Authentication required or invalid token
- `403 Forbidden`: Permission denied
- `404 Not Found`: Resource not found
- `405 Method Not Allowed`: HTTP method not supported
- `409 Conflict`: Resource already exists
- `500 Internal Server Error`: Server-side error
- `503 Service Unavailable`: Server temporarily unavailable

---

## Flutter Client Examples

### Complete Login Flow

```dart
class NulYunClient {
  final String baseUrl;
  String? _token;
  
  NulYunClient(this.baseUrl);
  
  Future<void> login(String username, String password) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/login'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'username': username,
        'password': password,
      }),
    );
    
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      if (data['otp'] == true) {
        // Handle TOTP - need to collect OTP code from user
        final otpCode = await getOTPCodeFromUser();
        await verifyOTP(data['token'], otpCode);
      } else {
        _token = data['token'];
      }
    } else {
      throw Exception('Login failed: ${response.body}');
    }
  }
  
  Future<void> verifyOTP(String tempToken, String code) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/login/otp'),
      headers: {
        'Content-Type': 'application/json',
        'X-Auth': tempToken,
      },
      body: jsonEncode({'code': code}),
    );
    
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      _token = data['token'];
    } else {
      throw Exception('OTP verification failed');
    }
  }
  
  Map<String, String> get _authHeaders => {
    'X-Auth': _token!,
  };
}
```

---

### List Files

```dart
Future<List<FileItem>> listFiles(String path) async {
  final response = await http.get(
    Uri.parse('$baseUrl/api/resources${Uri.encodeFull(path)}'),
    headers: _authHeaders,
  );
  
  if (response.statusCode == 200) {
    final data = jsonDecode(response.body);
    if (data['isDir'] == true) {
      return (data['items'] as List)
          .map((item) => FileItem.fromJson(item))
          .toList();
    }
    return [FileItem.fromJson(data)];
  }
  throw Exception('Failed to list files');
}

class FileItem {
  final String name;
  final int size;
  final DateTime modified;
  final bool isDir;
  final String? type;
  
  FileItem({
    required this.name,
    required this.size,
    required this.modified,
    required this.isDir,
    this.type,
  });
  
  factory FileItem.fromJson(Map<String, dynamic> json) {
    return FileItem(
      name: json['name'],
      size: json['size'],
      modified: DateTime.parse(json['modified']),
      isDir: json['isDir'],
      type: json['type'],
    );
  }
}
```

---

### Upload File

```dart
Future<void> uploadFile(File file, String destinationPath) async {
  var request = http.MultipartRequest(
    'POST',
    Uri.parse('$baseUrl/api/resources${Uri.encodeFull(destinationPath)}'),
  );
  
  request.headers.addAll(_authHeaders);
  request.files.add(await http.MultipartFile.fromPath(
    'file',
    file.path,
    filename: path.basename(file.path),
  ));
  
  final response = await request.send();
  if (response.statusCode != 200) {
    throw Exception('Upload failed');
  }
}
```

---

### Download File

```dart
Future<void> downloadFile(String filePath, String savePath) async {
  final response = await http.get(
    Uri.parse('$baseUrl/api/raw${Uri.encodeFull(filePath)}'),
    headers: _authHeaders,
  );
  
  if (response.statusCode == 200) {
    final file = File(savePath);
    await file.writeAsBytes(response.bodyBytes);
  } else {
    throw Exception('Download failed');
  }
}
```

---

### Search Files

```dart
Future<List<FileItem>> searchFiles(String query, {String? basePath}) async {
  final params = {
    'query': query,
    if (basePath != null) 'path': basePath,
  };
  
  final uri = Uri.parse('$baseUrl/api/search').replace(queryParameters: params);
  final response = await http.get(uri, headers: _authHeaders);
  
  if (response.statusCode == 200) {
    final data = jsonDecode(response.body);
    return (data['items'] as List)
        .map((item) => FileItem.fromJson(item))
        .toList();
  }
  throw Exception('Search failed');
}
```

---

### Create Share

```dart
Future<ShareInfo> createShare(
  String filePath, {
  String? password,
  DateTime? expires,
}) async {
  final body = {
    'path': filePath,
    if (password != null) 'password': password,
    if (expires != null) 'expires': expires.toIso8601String(),
  };
  
  final response = await http.post(
    Uri.parse('$baseUrl/api/share'),
    headers: {
      ..._authHeaders,
      'Content-Type': 'application/json',
    },
    body: jsonEncode(body),
  );
  
  if (response.statusCode == 200) {
    final data = jsonDecode(response.body);
    return ShareInfo(
      hash: data['hash'],
      url: data['url'],
    );
  }
  throw Exception('Failed to create share');
}

class ShareInfo {
  final String hash;
  final String url;
  
  ShareInfo({required this.hash, required this.url});
}
```

---

### Token Renewal

```dart
Future<void> renewToken() async {
  final response = await http.post(
    Uri.parse('$baseUrl/api/renew'),
    headers: _authHeaders,
  );
  
  if (response.statusCode == 200) {
    final data = jsonDecode(response.body);
    _token = data['token'];
  } else {
    throw Exception('Token renewal failed');
  }
}

// Check response headers after each request
void checkTokenRenewal(http.Response response) {
  if (response.headers['x-renew-token'] == 'true') {
    renewToken();
  }
}
```

---

## Best Practices

1. **Token Management**: Store tokens securely (e.g., `flutter_secure_storage` package). Check `X-Renew-Token` header and renew proactively.

2. **Error Handling**: Always handle network errors, timeouts, and server errors gracefully. Show user-friendly messages.

3. **Large File Uploads**: Use TUS protocol for files > 10MB. Implement progress tracking and resume capability.

4. **Caching**: Cache file listings and previews locally. Implement cache invalidation on file changes.

5. **Thumbnails**: Use `/api/preview/thumb/` for image thumbnails in lists to reduce bandwidth.

6. **Pagination**: For large directories, use `limit` and `offset` parameters to paginate results.

7. **WebDAV**: For advanced use cases (sync, mounting), use native WebDAV clients or libraries like `webdav_client` package.

8. **Security**:
   - Never log tokens
   - Use HTTPS in production
   - Implement certificate pinning for mobile apps
   - Clear tokens on logout

---

## Rate Limiting & Quotas

Currently, Nul Yun does not implement rate limiting. This may be added in future versions. Implement client-side rate limiting and retry logic with exponential backoff.

---

## Versioning

API version is indicated by the server version tag. Check `/api/settings` for server capabilities and feature availability.

---

## Support & Issues

For bugs, feature requests, or questions:
- GitHub: [nulnl/nulyun](https://github.com/nulnl/nulyun)
- Issues: Create an issue on GitHub

---

**Last Updated**: 2025-12-23  
**API Version**: v0.0.0-beta6
