VERDICT: CHANGES_REQUESTED

## Zusammenfassung

Geprüft wurde der vorliegende Go-Backend-Dienst vom Typ `go-backend`. Reine REST-API ohne Endnutzer-UI, daher entfallen Pflichttexte, Cookie-/Consent-Pflichten, EU AI Act und Barrierefreiheit. Relevant sind DSGVO und EU Cyber Resilience Act (CRA).

Positiv hervorzuheben:
- Body-Limit für POST und PUT funktioniert und liefert Status 413.
- Fehlerantworten sind generisch und enthalten keine Stacktraces, Dateipfade oder internen Go-Fehlermeldungen.
- Logging-Middleware erfüllt AC-13/AC-14: Es wird ausschließlich Methode, Pfad ohne Query-String und Statuscode geloggt.
- In-Memory-Store ist mit `sync.RWMutex` gegen Race Conditions geschützt.
- Die Nutzer-ID wird im Evaluate-Pfad nicht gespeichert und nicht geloggt.

Es bestehen jedoch behebbare Sicherheits- und Datenschutzlücken, insbesondere fehlende Authentifizierung, fehlender TLS-Schutz, fehlende CRA-Dokumentation und unzureichende Betriebshinweise für den Umgang mit personenbezogenen Eingaben. Kein fundamentaler Blocker, da keine PII im Klartext geloggt oder persistent gespeichert wird.

---

## 1. EU Cyber Resilience Act (CRA)

### CRA-1 — Hoch — Fehlende Authentifizierung und Autorisierung
**Datei:** `main.go`, `internal/api/middleware.go`

Sämtliche Endpunkte außer `/healthz` sind ungeschützt. Jeder, der Netzwerkzugriff auf den Dienst hat, kann Flags anlegen, ändern, löschen und Evaluate-Aufrufe mit beliebigen Nutzer-IDs ausführen. Das verletzt Security by Design/Default (CRA) und gefährdet Integrität und Vertraulichkeit.

**Maßnahme:**
- Middleware für API-Key/Bearer-Token implementieren, z. B. `RequireAPIKey(key string)` oder `RequireBearerToken`.
- Mindestens POST/PUT/DELETE schützen; Evaluate je nach Einsatz ebenfalls.
- Routen in `main.go` entsprechend wrappen.

---

### CRA-2 — Hoch — Keine Transportverschlüsselung (TLS)
**Datei:** `main.go`

Der Server startet mit `http.ListenAndServe` auf `:8080` ohne TLS. Werden Flag-Daten oder die Nutzer-ID über ein Netz übertragen, geschieht dies im Klartext. Das verletzt CRA-Security-Defaults und DSGVO Art. 32.

**Maßnahme:**
- Entweder `server.ListenAndServeTLS("cert.pem", "key.pem")` verwenden oder
- einen vorgeschalteten TLS-Proxy verpflichtend dokumentieren (z. B. in `README.md`).
- Zusätzlich `http.Server.TLSConfig` und sichere TLS-Mindestversionen konfigurieren.

---

### CRA-3 — Mittel — Fehlende Server-Timeouts
**Datei:** `main.go`

Der `http.Server` hat keine `ReadTimeout`, `ReadHeaderTimeout`, `WriteTimeout` oder `IdleTimeout`. Langsame Clients können Verbindungen lange blockieren und Ressourcen erschöpfen.

**Maßnahme:**
```go
server := &http.Server{
    Addr:              ":8080",
    Handler:           handler,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       60 * time.Second,
}
```
`import "time"` in `main.go` ergänzen.

---

### CRA-4 — Mittel — Fehlende SBOM und Sicherheitsdokumentation
**Datei:** fehlende `SBOM.md`/`sbom.json` und `SECURITY.md`

Im Branch ist keine SBOM- oder Security-Dokumentation sichtbar. Der CRA verlangt für Produkte mit digitalen Elementen eine Software Bill of Materials, dokumentierte Sicherheitsanforderungen, einen Schwachstellenmeldeprozess und Update-/Patch-Fähigkeit.

**Maßnahme:**
- `SBOM.md` oder `sbom.json` (CycloneDX/SPDX) mit Modul `featureflags`, Go-Version und Abhängigkeiten (hier: nur Standardbibliothek) anlegen.
- `SECURITY.md` mit Kontakt für Schwachstellenmeldungen, Disclosure-Policy und Update-Prozess ergänzen.

---

### CRA-5 — Mittel — Keine Rate-Limitierung
**Datei:** `main.go` oder `internal/api/middleware.go`

Der Evaluate-Endpoint kann unbegrenzt mit unterschiedlichen `user`-Werten aufgerufen werden. Ohne Auth und Rate-Limit erleichtert das Missbrauch und unerwünschte Auswertungen.

**Maßnahme:**
- Rate-Limit-Middleware (z. B. Token-Bucket pro Client-IP/API-Key) vor die kritischen Routen schalten.
- Limits in Betriebsdokumentation konfigurierbar machen.

---

### CRA-6 — Niedrig — Härtungs-Header für JSON-Antworten
**Datei:** `internal/api/response.go`

Es wird nur `Content-Type: application/json` gesetzt. Für eine JSON-API empfiehlt sich zumindest `X-Content-Type-Options: nosniff`.

**Maßnahme:**
- In `WriteJSON` zusätzlich `w.Header().Set("X-Content-Type-Options", "nosniff")` setzen.
- Optional `Cache-Control: no-store` für flag- und evaluate-Antworten.

---

## 2. DSGVO

### DSGVO-1 — Mittel — Nutzer-ID im Query-String und ohne TLS
**Datei:** `internal/api/flags_evaluate.go`, `main.go`

`EvaluateFlag` liest `r.URL.Query().Get("user")`. Die eigene Logging-Middleware loggt den Pfad ohne Query-String (positiv, AC-14 erfüllt). Vorgelagerte Systeme, Proxies oder externe Access-Logs können den Query-String jedoch erfassen. In Kombination mit fehlendem TLS (CRA-2) ist die Vertraulichkeit der Nutzer-ID nicht ausreichend geschützt.

**Maßnahme:**
- TLS wie bei CRA-2 umsetzen.
- In `README.md` oder `AGENTS.md` verbindlich festlegen, dass `user` ausschließlich eine pseudonyme, nicht direkt rückführbare ID sein darf (keine E-Mail, Telefonnummer, Personalnummer o. Ä.).
- Optional: künftige API-Version auf POST mit Body oder Header-Feld umstellen, falls der API-Vertrag das zulässt.

---

### DSGVO-2 — Mittel — Rechtsgrundlage und Verarbeitungsdokumentation nicht sichtbar
**Datei:** `README.md` / `AGENTS.md`

Der Dienst verarbeitet im Evaluate-Pfad Nutzer-IDs transient. Eine dokumentierte Rechtsgrundlage (Art. 6 DSGVO), der Verarbeitungszweck und ein Verantwortlicher sind im vorgelegten Code nicht sichtbar. Da `README.md` existiert, aber sein Inhalt nicht Teil des Reviews war, ist dies nicht abschließend prüfbar.

**Maßnahme:**
- Datenschutzabschnitt im `README.md` oder `AGENTS.md` ergänzen:
  - Zweck: deterministische Feature-Flag-Auswertung.
  - Rechtsgrundlage: z. B. berechtigtes Interesse Art. 6 Abs. 1 lit. f DSGVO oder Auftragsverarbeitungsvereinbarung je nach Einsatz.
  - Datenarten: undurchsichtige `user`-IDs.
  - Speicherdauer: keine Persistenz der Nutzer-ID.
  - Betroffenenrechte: Hinweis, wie Anfragen gestellt werden können.

---

### DSGVO-3 — Mittel — `description` ohne Nutzungsbeschränkung und ohne Löschkonzept
**Datei:** `internal/api/flags_create.go`, `internal/api/flags_update.go`, `internal/store/store.go`

POST und PUT erlauben beliebige `description`-Strings bis zum Body-Limit. Beschreibungen werden im In-Memory-Store bis DELETE oder Neustart gespeichert. Enthalten sie personenbezogene Daten, fehlt Datenminimierung und ein klar dokumentiertes Löschkonzept.

**Maßnahme:**
- In `README.md` oder API-Doku verbindlich festlegen, dass `description` ausschließlich technische Metadaten ohne Personenbezug enthalten darf.
- Optional: Validierung auf offensichtliche PII-Muster oder striktere Längenbegrenzung für `description` einführen.

---

### DSGVO-4 — Niedrig — Kein expliziter Löschhinweis für mögliche PII in Flag-Daten
**Datei:** `README.md`

`DELETE /flags/{key}` existiert und entfernt das Flag, aber es fehlt eine Betriebsanweisung, was bei Verdacht auf personenbezogene Daten in Beschreibungen zu tun ist.

**Maßnahme:**
- Im Datenschutzabschnitt des `README.md` einen Hinweis aufnehmen, dass verdächtige Flag-Daten unverzüglich via DELETE entfernt werden müssen.

---

## 3. EU AI Act

Kein KI-Feature sichtbar. Keine Pflichten.

---

## 4. Pflichttexte & UI

Reines Backend ohne Endnutzer-UI. Keine Legal-Notice-, Cookie-/Consent- oder AGB-Pflichten. Keine Befunde.

---

## 5. Barrierefreiheit

Keine öffentliche Web-UI vorhanden. Nicht anwendbar.

---

## Gesamteinschätzung

Behebbare Lücken bei Authentifizierung, TLS, Server-Härtung und CRA-Dokumentation sowie ergänzungsbedürftige Datenschutzhinweise. Es wurden keine personenbezogenen Daten im Klartext geloggt oder persistent gespeichert. Daher kein fundamentaler Verstoß, aber vor Marktfreigabe sind die genannten Maßnahmen umzusetzen.