# BYTE SIZE

Content-addressable storage with chunk-level deduplication, exposed via a REST API.  
Built with **Go**, **PostgreSQL**, and designed to scale with **goroutines**, **channels**, and (later) **Docker/K8s**.

---

## Features
- **File ingestion via REST API** (`/files/upload`).
- **Gets a certain File MetaData** (`/files/metadata/:id`)
- **File download via REST API** (`/files/download/:id`) — streams the reconstructed file from chunks.
- Automatic chunking (default: 4 MiB per chunk).
- SHA-256 content hashing and deduplication.
- Persistent chunk storage on disk (`FSChunkStore`).
- PostgreSQL-backed metadata:
  - Files
  - Chunks
  - File–chunk manifests
- Transactional batch inserts with rollback on failure.
- Structured error handling and JSON responses.

---

## Auth
All routes are protected by an API key.

- Header: `X-API-Key: <your-key>`
- or Query: `?api_key=<your-key>`

`MIDDLEWARE_KEY` is read from environment.

---

## Quick API Examples

```bash
# Upload (multipart/form-data)
curl -H "X-API-Key: $MIDDLEWARE_KEY" -F "file=@/path/to/file.bin" \
  http://localhost:8080/files/upload

# Get metadata
curl -H "X-API-Key: $MIDDLEWARE_KEY" \
  http://localhost:8080/files/metadata/<uuid>

# Download (respects Content-Disposition filename)
curl -H "X-API-Key: $MIDDLEWARE_KEY" -OJ \
  http://localhost:8080/files/download/<uuid>

# Metrics (Prometheus text format)
curl -H "X-API-Key: $MIDDLEWARE_KEY" \
  http://localhost:8080/metrics
```

## Tech Stack
- Go (concurrency + service layer)
- PostgreSQL (metadata storage)
- Local FS (FSChunkStore) for chunk data 
- httprouter for routing 
- validator.v10 for request validation

## Contributing
- Will be developed alone

⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⢟⣫⣽⣶⣶⣶⣆⠠⣉⣴⣾⣿⣿⣿⣷⣶⠈⠍⠙⠿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠃⣾⣿⣿⣿⣿⣿⣿⡌⠘⡿⠟⠻⢿⣿⣿⠿⠟⢸⣿⣿⡇⠁⢾⣿⣿⣿⣿⣿⡞⣿⣿⡿⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⣫⠛⠋⣠⣾⣿⣿⣿⢡⣻⡏⢣⠀⣷⡄⣴⣿⣿⢣⡆⠸⢸⣿⡿⠁⣰⣷⣦⡌⢿⣿⣿⣿⢸⣿⠇⠿⢿⠟⠋⣩⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⢋⣥⣾⣿⣿⣷⣦⡹⣿⣿⣿⠘⡧⡇⠈⠀⡿⠟⠘⢛⣛⣘⡀⣘⣿⣟⣀⡀⠿⢿⢛⢅⡤⣿⣿⣿⡇⡿⠀⠀⠀⠠⢌⣒⣶⣖⡂⢩⣽⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⣴⣿⣿⣎⢿⣿⡿⢻⣧⠰⣾⣁⠀⣸⣷⡨⠐⠒⣒⡭⠅⡀⠙⠻⢶⣮⣝⡫⢙⢿⣶⡶⢦⣀⠉⢿⣿⡇⠁⠀⢀⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⠡⣾⣿⣿⣿⣿⣆⠹⣀⢘⡛⠃⠉⢀⠔⣩⡿⠁⢻⣶⣤⣵⣄⢦⢡⡠⢄⠙⢿⣿⣶⣅⠙⢷⣦⡙⠳⣄⠙⢁⡀⢀⣡⣤⣤⣤⡈⠙⠻⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣧⢻⣿⠿⣿⣿⣿⣦⠘⢈⠜⠁⠀⣠⣾⣿⠃⢠⣽⣏⢻⣿⣿⣧⡂⢿⣷⣄⠱⣽⣿⣿⣿⣆⠹⡹⣦⣀⠀⠀⠐⢎⣉⠛⢿⣿⣿⣷⡑⢻⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⢃⣿⡐⢮⡙⢛⣛⡃⠀⠀⠀⣰⣿⣿⠃⠀⣼⣟⣿⡇⢻⣿⣿⣷⡈⢿⣿⡛⣷⠹⣿⣿⣿⣧⡀⢻⢻⣷⣄⠀⠈⢻⣿⡆⡍⠉⠙⢿⣌⣿⣿⣿⣿⣿⣿⣿
⡙⣿⣿⣿⣿⣿⣿⡟⣡⡏⠀⠀⠀⢨⣂⠙⠑⢀⡄⢀⣿⣿⣿⠀⡄⣿⢸⣿⡟⡄⢹⣿⣿⣷⡘⣿⡇⢸⣇⠹⣿⣿⣿⣿⠀⣧⣹⣿⣷⣄⠀⠀⢤⡀⠀⠀⠈⢿⣿⣿⣿⣿⣿⣿⣿
⣿⣮⠻⣿⣿⣿⣟⠀⣿⣿⣶⣾⣿⡆⠿⠁⢠⡾⠁⣸⣿⣿⡇⢰⢿⣿⢸⣿⣿⢰⣠⣿⣿⣿⣧⣸⣷⠀⣇⠀⢹⣿⣿⢹⠀⢸⣧⠡⡹⣿⣷⠑⠌⢻⣾⣶⣄⠈⠹⣿⣿⣿⣿⣿⣿
⣿⣿⣷⣝⠻⣿⣿⣇⣿⣿⣿⡟⢋⣠⠀⢠⢟⡼⠃⣿⣿⣿⡇⢸⢸⣿⠈⣿⣿⡈⣿⡏⢻⣿⣿⡇⢿⠀⣿⠀⠈⣿⣿⡸⠀⠀⢻⣇⢱⠙⣿⣇⠈⢆⠉⢉⣛⡃⠠⠘⢿⣿⣿⣿⣿
⣄⠉⠙⠻⢦⡈⢻⣿⡘⠿⠋⠔⠛⠋⠀⢏⠞⣱⠀⣿⣿⣿⡇⢸⢸⡟⠀⣿⣿⡇⢹⣇⠀⢻⣿⣿⣸⠀⣿⠀⠀⢸⣿⡇⠀⠀⠘⣿⡎⠀⠛⣿⠀⠘⡆⠀⠙⠻⣿⣶⣤⣙⡻⣿⣿
⣭⣭⣄⡀⠀⠈⠀⠙⣿⣆⠀⠀⠠⠃⢰⠊⣼⠏⢠⣿⣿⣿⣇⠀⢸⡇⠀⣿⣿⡇⣿⣿⠀⢆⢻⣿⡇⠀⣿⠀⠀⢸⡿⠀⠾⠞⠂⠛⣣⣦⣆⢹⠀⠀⠸⣦⠀⢀⠀⠀⠀⣀⠀⣿⣿
⠭⠭⣛⣛⠻⠶⢤⡀⠈⠻⠀⠀⠀⠀⠈⠞⢡⡆⢰⣿⣿⣿⣿⡀⢺⣿⠀⣿⣿⡇⣿⣿⢠⠘⡌⣿⠇⣠⠿⣀⣵⢆⠏⠀⠀⠨⣭⣭⣑⡒⠀⠈⠁⠀⠂⢿⣇⠀⠀⠉⠂⠀⠁⣿⣿
⠿⠗⣒⣂⣤⣤⠄⠀⠀⠀⠀⠀⠀⡀⣀⣴⣿⡇⠈⣿⡿⣿⣿⡇⢸⣿⠀⣿⣿⢧⢹⡿⠼⢀⣁⣙⠺⢶⣿⣿⣿⡸⣤⡀⠠⢰⢸⣿⡿⠇⢶⠠⡀⠀⠀⢸⣿⠀⠀⠀⠀⠀⠀⠸⣿
⡶⢖⣒⣭⡥⠖⠋⠁⠀⠀⠷⠐⠀⠘⣿⣿⣿⣇⢀⢹⣿⠘⣿⣧⠘⣿⠀⣿⡿⠸⠈⠁⠉⣐⣤⣀⢐⣛⣿⣿⣿⣷⣬⣿⣶⣤⣭⡁⣄⣠⣴⡆⠃⠀⠀⢸⣿⠀⢀⠀⠀⠀⠀⠀⢹
⣶⡿⠿⠁⠀⠀⠀⠀⠀⠀⣁⠀⠠⡁⣦⣭⣭⣿⡘⡎⣿⣧⠘⣿⣇⠹⠀⣿⠀⠋⠀⠐⠛⢿⣿⡿⠃⣿⣿⣿⣿⢰⠖⣩⡏⢿⣿⣿⣿⣿⣿⣿⡆⠀⠀⢸⠏⢷⡰⣄⠀⠀⠀⠀⠘
⡇⠀⠀⠀⠀⠀⠀⡀⠀⡀⢻⡷⢰⣧⠘⠿⠟⠋⠁⠻⠘⣿⣧⡘⣿⣄⠡⠀⠀⡀⠆⡭⢨⢸⠟⢁⣾⣿⣿⣿⣿⣶⣾⡟⣿⢘⣿⣿⣿⣿⣿⣿⣧⠀⠀⣸⠀⠈⠃⠘⢿⣦⠀⢀⠀
⠁⠀⠀⠀⠀⠀⣰⣿⣿⡇⠘⢇⣿⣿⠀⡈⠻⣿⣿⡀⢷⠹⣿⣇⠘⠖⡀⠂⠸⠿⠦⠀⡒⢠⣴⣿⣿⣿⣿⣿⣿⣿⣿⣷⣿⣾⡿⠿⣿⣿⣿⣿⡏⠁⠈⠉⠁⠀⢢⣄⣀⡙⠳⢤⣄
⠀⠀⠀⠂⠀⣾⣿⣿⣿⠀⣶⣄⢻⠃⠀⠿⠶⠌⠻⠧⠈⣦⠹⣿⣷⣄⢥⣦⡀⢿⣄⣠⣷⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠩⠒⠀⣼⣿⣿⣿⡇⠀⠀⠀⠀⠀⢸⣿⣿⣿⣶⣶⣶
⠀⠀⠀⢀⣼⣿⣿⣿⣿⡇⠘⣧⡀⢸⡀⢀⠉⠛⣛⡛⡀⣌⢧⠘⣿⣿⣷⣝⠻⣄⠈⢛⠻⠿⣿⣿⣿⣿⣿⣿⣿⣿⡟⠁⠀⠀⠀⠀⣿⣿⣿⣿⢣⡆⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⣿
⠠⠀⡇⡄⣿⣿⣿⣿⣿⢣⣷⡜⠃⣼⡧⢈⡳⣦⠀⠴⠶⠸⣎⢣⡈⢿⣿⣿⣷⣍⠳⠆⠍⠓⠶⠾⢿⣿⣿⣿⣿⡟⠀⠀⠀⠀⠀⠀⣿⣿⣿⡏⣼⠁⠀⣠⠀⠀⣸⣿⣿⣿⣿⣿⣿
⡸⠃⠐⠀⣿⣿⣿⣿⣿⢸⣿⣷⠠⣿⡇⠀⠻⣦⡌⡁⣠⠄⢹⡆⠑⢄⠻⣿⣿⣿⣷⣤⡀⠒⢶⣶⣶⣿⣿⣿⣿⡇⢀⣤⣄⠑⢦⡀⣿⣿⠟⣸⣧⣾⣾⣿⠀⣰⣿⣿⣿⣿⣿⣿⣿
⢀⣤⠀⢀⣿⣿⣿⡿⠋⡐⣿⣿⡇⠛⣠⣷⣄⠐⢥⡀⠹⣈⣀⢻⣆⢮⡢⡙⢿⣏⡛⠿⠿⠷⠴⣭⡙⠛⠿⣿⣿⣿⡌⠙⣿⣷⣮⣇⢻⠟⡰⣿⣿⣿⣿⣿⣼⣿⣿⣿⣿⣿⣿⣿⣿
⣴⣿⡆⠶⣬⢻⣿⣄⢰⣿⣮⠻⣧⣤⠉⠻⠿⠿⠦⠍⠀⠙⠃⠀⢿⡷⠙⢮⡀⠙⢿⣄⠐⢶⣶⣶⣶⣿⣿⣿⣿⣿⣿⣦⣈⣙⡛⢿⠘⠀⠀⢬⡙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⡄⢴⣿⣿⠶⠘⠏⣴⢸⣿⡿⠱⠞⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣇⡆⠀⠉⠂⠀⠙⢦⡈⠹⣆⠀⠈⠉⠛⠛⠻⠿⠿⢿⣿⣿⡗⠀⠀⠀⠡⡄⢳⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣶⡄⣰⣿⣿⣷⣿⡆⠟⣠⢀⣴⣶⣆⢻⣶⣦⣄⡀⠀⠀⠀⠀⠀⠈⠀⠀⠀⠀⠐⠥⣈⠳⢄⠙⢷⣄⠀⠀⣀⣤⡶⠀⠀⠀⠀⣀⡀⠀⠀⠀⠈⢀⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣷⣮⣍⣉⠈⢋⣚⣛⠁⢸⣿⣿⢿⣦⠹⣿⣿⣿⣿⣶⣤⣴⣤⡄⢀⠀⠀⢀⡀⢻⣦⣤⡀⠀⢀⣈⣛⠿⣿⡟⣁⠀⠀⠈⣎⢿⢇⡄⢿⡐⢀⠈⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⡈⢹⣷⣄⡙⠧⢹⣿⣿⣿⣿⣿⣿⣿⡇⠟⣡⣶⣿⡇⠘⣻⢿⣷⣌⠲⣦⣽⣿⣾⡇⠻⣷⣦⡀⢸⡎⣼⡇⢸⣿⡏⠄⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢰⣿⣿⣿⣦⡘⢿⣿⣿⣿⣿⣿⠟⣡⣾⣿⣿⣿⣿⣄⠻⡀⠙⠻⣷⣌⡛⠿⣿⣿⣦⣽⣿⣿⢸⣷⠸⡇⢸⣿⣿⡄⠂⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡏⠘⣩⣵⡶⣶⡆⠀⠹⣿⣿⣿⠇⣼⣿⣿⣿⣋⣉⣩⣿⣧⡈⠀⢐⢮⡛⢿⣶⣤⣉⣛⣿⠿⠏⠼⠿⡇⢃⣸⣿⣿⣷⡄⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⡜⢛⣅⣴⣿⣧⡉⠠⡀⠻⢃⡾⠟⣛⠻⠯⠩⠭⠭⣽⣿⣿⣦⡀⠁⠙⢷⣬⡛⢿⣿⣿⢋⣄⠀⠀⠈⠄⢿⣿⣿⣿⣷⠈⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠻⢠⣾⣿⡿⠟⡩⠶⠛⠓⠈⢠⡀⠈⠛⠿⢿⣿⣿⣿⣷⣬⣲⣾⣿⣿⣆⠀⠙⢿⣿⣿⣿⡏⠀⠀⠉⠀⠀⢠⡀⠻⣿⣿⣿⠀⢿⣿⣿⣿⣿⣿⣿⣿⣿