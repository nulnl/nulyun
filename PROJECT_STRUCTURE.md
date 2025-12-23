# Project Structure

This document describes the project's refactored directory structure.

# Backend (Go)

The project follows the Go standard project layout:

```
├── cmd/                      # Application entry points
│   └── nulyun/              # Main application
│       └── main.go          # Program entry point
│
├── pkg/                     # Public libraries safe for external import
│   └── version/             # Version information
│
├── internal/                # Private application code (not for external import)
│   ├── auth/               # Authentication module (JSON, Proxy, NoAuth)
│   ├── files/              # File handling services
│   ├── handler/            # HTTP request handlers
│   ├── middleware/         # HTTP middleware
│   ├── model/              # Data models
│   │   ├── global/        # Global settings models
│   │   ├── passkey/       # Passkey/WebAuthn models
│   │   ├── share/         # Sharing models
│   │   ├── users/         # User models
│   │   └── webdav/        # WebDAV models
│   ├── pkg_errors/         # Error definitions
│   ├── repository/         # Data access layer
│   │   └── bolt/          # BoltDB storage implementation
│   ├── service/            # Business logic layer
│   └── version/            # Version information (to be removed)
│
├── www/                     # Frontend assets (embedded into the binary)
│   └── assets.go           # Go file embedding frontend assets
│
├── docs/                    # Documentation
│   ├── API.md              # API documentation
│   └── INTEGRATION.md      # Integration guide
│
├── bin/                     # Compiled binaries
│   └── nulyun              # Main binary
│
├── go.mod                   # Go module definition
├── go.sum                   # Go dependency checksums
├── Makefile                 # Build scripts
├── Dockerfile               # Docker image build
└── README.md                # Project documentation
```

### Directory Responsibilities

#### cmd/
- Contains application entry points
- Each subdirectory builds an executable
- Keep the command entry minimal; most logic lives under `internal/`

#### pkg/
- Public libraries that can be safely imported by external projects
- Currently contains shared utilities like version information

#### internal/
Core private application code that should not be imported by external projects:

- **auth/**: Authentication and authorization logic
- **files/**: File operations, caching, and image processing
- **handler/**: HTTP routing and request handling
- **model/**: Data models and business entities
- **repository/**: Data persistence layer and data access interfaces
- **service/**: Business logic services
- **pkg_errors/**: Custom error types

## Frontend (Vue.js + TypeScript)

```
www/
├── src/
│   ├── api/                 # API clients
│   │   ├── files.ts        # Files API
│   │   ├── users.ts        # Users API
│   │   ├── settings.ts     # Settings API
│   │   ├── share.ts        # Share API
│   │   ├── passkey.ts      # Passkey API
│   │   ├── webdav.ts       # WebDAV API
│   │   └── ...
│   │
│   ├── components/          # Vue components
│   │   ├── files/          # File-related components
│   │   ├── header/         # Header components
│   │   ├── prompts/        # Modal/prompt components
│   │   ├── settings/       # Settings components
│   │   └── ...
│   │
│   ├── views/               # Page views
│   │   ├── Files.vue       # File manager page
│   │   ├── Settings.vue    # Settings page
│   │   ├── Login.vue       # Login page
│   │   ├── Share.vue       # Share page
│   │   └── ...
│   │
│   ├── stores/              # Pinia state management
│   │   ├── auth.ts         # Authentication state
│   │   ├── file.ts         # File state
│   │   ├── layout.ts       # Layout state
│   │   ├── upload.ts       # Upload state
│   │   └── ...
│   │
│   ├── router/              # Router configuration
│   ├── i18n/                # Internationalization
│   ├── types/               # TypeScript type definitions
│   ├── utils/               # Utility functions
│   ├── css/                 # Stylesheets
│   ├── assets/              # Static assets
│   ├── App.vue              # Root component
│   └── main.ts              # Entry file
│
├── public/                  # Public static files
├── dist/                    # Build output (embedded into the Go binary)
├── package.json             # npm package configuration
├── vite.config.ts           # Vite build configuration
└── tsconfig.json            # TypeScript configuration
```

### Frontend Architecture

- **SPA**: Uses Vue Router for client-side routing
- **State Management**: Pinia for global state
- **Type Safety**: TypeScript for type checking
- **Modularity**: APIs, components, and views organized by feature
- **Internationalization**: Supports multiple languages

## Build Process

### Development

```bash
# Backend development
make build-backend

# Frontend development
cd www && pnpm run dev

# Full build
make build
```

### Production

```bash
# Build frontend and embed into the Go binary
make build

# Build Docker image
docker build -t nulyun .
```

## Key Improvements

### Backend refactor
1. ✅ Follows Go standard project layout
2. ✅ Clear layered architecture (Handler -> Service -> Repository)
3. ✅ `main.go` moved to `cmd/nulyun`
4. ✅ Improved package names (`http` -> `handler`, `errors` -> `pkg_errors`, `settings` -> `model`, `storage` -> `repository`)
5. ✅ Clearer module responsibilities

### Frontend improvements
1. ✅ Retains modern Vue 3 + TypeScript stack
2. ✅ Reasonable directory structure
3. ✅ Modular API clients
4. ✅ Components grouped by feature

### Configuration updates
1. ✅ `Makefile` updated with new build path
2. ✅ `Dockerfile` updated entry point
3. ✅ `.goreleaser.yml` updated main path

## Further Improvement Suggestions

1. Consider removing `internal/version` and standardizing on `pkg/version`
2. Further split large files in `handler/` into smaller units
3. Add unit test coverage
4. Consider introducing a dependency injection tool (e.g., `wire`)
5. Further optimize frontend bundle size

