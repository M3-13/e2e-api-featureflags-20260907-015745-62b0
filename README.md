# Feature-Flag-Service (Go)

Ein Feature-Flag-Service als REST-API in reinem Go (nur Standardbibliothek
`net/http`). Flags lassen sich anlegen, listen, abrufen, ändern und löschen;
pro Nutzer liefert ein Evaluate-Endpunkt eine deterministische
Ja/Nein-Entscheidung basierend auf einem stabilen Hash aus Flag-Key und
Nutzer-ID gegen `rollout_percent`. Thread-sicherer In-Memory-Store,
Eingabevalidierung mit sauberen Statuscodes und JSON-Fehlerobjekten,
Zugriffs-Logging als Middleware sowie Go-Tests mit `httptest` für jeden
Handler und die Rollout-Entscheidung.

## Tech-Stack

- **Sprache**: Go (Standardbibliothek `net/http`)
- **Storage**: In-Memory mit `sync.RWMutex`
- **Testing**: Go `testing` + `net/http/httptest`

## Installation & Start

```sh
# Build prüfen
go build ./...

# Dienst starten (Port 8080)
go run .

# Tests ausführen
go test ./...
```

Der Dienst lauscht auf `:8080`.

## Endpunkte

| Methode | Pfad                    | Beschreibung                              |
|---------|-------------------------|-------------------------------------------|
| POST    | `/flags`                | Flag anlegen                               |
| GET     | `/flags`                | Alle Flags (nach key sortiert) listen      |
| GET     | `/flags/{key}`          | Einzelnes Flag abrufen                     |
| PUT     | `/flags/{key}`          | Flag ändern (enabled, description, rollout) |
| DELETE  | `/flags/{key}`          | Flag löschen                               |
| GET     | `/flags/{key}/evaluate` | Rollout-Entscheidung pro Nutzer (`?user=`) |
| GET     | `/healthz`              | Health-Check (`{"status":"ok"}`)           |

Alle Antworten sind JSON; Fehler haben die Form `{"error":"..."}`.
