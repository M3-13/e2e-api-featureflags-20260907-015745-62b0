# SBOM – Software Bill of Materials

## Modul

- **Name:** `featureflags`
- **Typ:** Go-Modul (REST-API)

## Go-Version

- Go 1.22

## Abhängigkeiten

Der Dienst verwendet ausschließlich Pakete der Go-Standardbibliothek. Es bestehen
keine Third-Party-Abhängigkeiten (kein `require`-Block in `go.mod`).

| Paket           | Zweck                                                              |
|-----------------|--------------------------------------------------------------------|
| `net/http`      | HTTP-Server, Routing, Handler und Middleware                       |
| `encoding/json` | JSON-Serialisierung und -Deserialisierung                          |
| `sync`          | Thread-sichere Synchronisation (`sync.RWMutex`)                    |
| `hash/fnv`      | FNV-1a-Hash für die deterministische Rollout-Entscheidung          |
| `log`           | Zugriffs- und Fehler-Logging                                       |
| `regexp`        | Validierung des Feature-Flag-Key-Formats                           |
| `crypto/subtle` | Konstante-Zeit-Vergleiche für den API-Key                          |
| `time`          | Zeitstempel und Timeouts                                           |
