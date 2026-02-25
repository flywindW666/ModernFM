# ModernFM Docker Version Architecture

## 1. Stack Overview (Docker Compose)
- **Backend Service (Go)**: The core API server.
- **Frontend Service (Nginx/Vue)**: Serving the UI assets.
- **Database (PostgreSQL)**: Storing file metadata, user permissions, and sharing links.
- **Cache (Redis)**: Caching directory listings and session data for ultra-fast response.

## 2. Technical Changes
- **Metadata Management**: Instead of raw `os.ReadDir`, we now scan and index files into **PostgreSQL**.
- **Performance**: Directory structures are cached in **Redis** with TTL.
- **Persistence**: File data remains on the host (e.g., Unraid array), mounted via Docker Volumes.

## 3. Database Schema (PostgreSQL)
- `files`: id, name, path, parent_id, size, mod_time, is_dir, hash.
- `users`: id, username, password_hash, role.
- `shares`: id, file_id, token, expire_at, password.
