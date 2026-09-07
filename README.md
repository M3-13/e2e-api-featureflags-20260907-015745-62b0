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

## Datenschutz

### Zweck

Der Dienst dient ausschließlich der **deterministischen Feature-Flag-Auswertung**:
für einen gegebenen Flag-Key und eine Nutzer-ID wird reproduzierbar entschieden, ob
ein Feature für diesen Nutzer aktiviert ist. Es findet kein Profiling, Tracking
oder anderweitige Auswertung des Nutzerverhaltens statt.

### Rechtsgrundlage

Die Verarbeitung erfolgt auf Grundlage von **Art. 6 Abs. 1 lit. f DSGVO**
(berechtigtes Interesse). Das berechtigte Interesse liegt in der technischen
Notwendigkeit, Feature-Flags kontrolliert und reproduzierbar auszurollen, ohne
dass hierfür personenbezogene Merkmale erforderlich sind.

### Datenarten

- **Undurchsichtige, pseudonyme user-IDs** als Query-Parameter `?user=` beim
  Evaluate-Endpunkt.
- **Keine** E-Mail-Adressen, Telefonnummern, Personalnummern oder sonstige direkt
  identifizierende Merkmale werden verarbeitet oder gespeichert.
- Der Wert `user` darf ausschließlich pseudonym sein und keine personenbezogenen
  Daten enthalten.
- Das Feld `description` eines Flags darf keine personenbezogenen Daten enthalten.

### Speicherdauer

Die Nutzer-ID wird **nicht persistiert**. Sie fließt lediglich transient in die
Hash-Berechnung (FNV-1a) ein und wird danach verworfen. Eine Speicherung oder
Zuordnung von Nutzer-IDs findet zu keinem Zeitpunkt statt.

### Betroffenenrechte

Da keine personenbezogenen Daten gespeichert werden, sind die Betroffenenrechte
(Auskunft, Berichtigung, Löschung, Einschränkung, Datenübertragbarkeit) insoweit
gegenstandslos. Für Rückfragen wenden sich Betroffene an die in `SECURITY.md`
genannte Kontaktstelle.

### Verdächtige Flag-Daten

Verdächtige Flag-Daten – insbesondere personenbezogene Daten, die entgegen dieser
Vorgaben in `description` oder `key` auftauchen – sind unverzüglich über
`DELETE /flags/{key}` zu entfernen.

## Betrieb

### API-Key

Für geschützte Routen muss die Umgebungsvariable `API_KEY` gesetzt sein. Ohne
gültigen `API_KEY` antworten geschützte Routen mit **503**.

### TLS

Die Transportverschlüsselung erfolgt entweder über die Umgebungsvariablen
`TLS_CERT_FILE` (Zertifikat) und `TLS_KEY_FILE` (privater Schlüssel) oder über
einen vorgeschalteten, TLS-terminierenden Reverse-Proxy.

### Rate-Limit

Die Anfragefrequenz wird über die Umgebungsvariable `RATE_LIMIT_PER_MINUTE`
begrenzt. Ohne gesetzten Wert gilt ein **Default von 120** Anfragen pro Minute.
