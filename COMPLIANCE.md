VERDICT: CHANGES_REQUESTED

## Gesamtbewertung

Der Dienst setzt viele Security-Baselines bereits sauber um: generische JSON-Fehler ohne interne Details, Body-Limits für POST/PUT, Logging ohne Query-String, Constant-Time-Vergleich für den API-Key, Server-Timeouts, optionale TLS-Konfiguration mit `MinVersion: TLS1.2`, Rate-Limiting sowie ein threadsicherer Store. Als reines `go-backend` ohne Endnutzer-UI bestehen keine Impressums-, Cookie-, Datenschutzerklärungs- oder Barrierefreiheitspflichten im Code.

Offen sind zwei behebbare Mängel: eine unbefristete Speicherung von IP-Adressen im Rate-Limiter sowie ein unverschlüsselter HTTP-Standardbetrieb. Beides ist kein fundamentaler Blocker, muss aber vor Auslieferung behoben werden.

---

## 1. DSGVO

### Befund DSGVO-1 — Unbefristete IP-Speicherung im Rate-Limiter (hoch)

**Datei:** `internal/api/ratelimit.go`

Die Middleware speichert für jeden Client einen Eintrag in der Map `clients`, identifiziert über `r.RemoteAddr`. `r.RemoteAddr` enthält die IP-Adresse und den Port des Clients und ist damit ein personenbezogenes Datum. Die Map wird niemals bereinigt und besitzt keine Obergrenze.

Das verletzt:

- Art. 5 Abs. 1 lit. c DSGVO — Datenminimierung, da IP-Adressen ohne Begrenzung gesammelt werden.
- Art. 5 Abs. 1 lit. e DSGVO — Speicherbegrenzung, da es keine Löschroutine oder Aufbewahrungsfrist gibt.

Zusätzlich entsteht ein praktischer DoS-Vektor: Ein Angreifer kann mit vielen unterschiedlichen Quell-IP-Adressen die Map unbegrenzt wachsen lassen.

**Remedy:**

- In `internal/api/ratelimit.go` eine TTL einführen, z. B. 5 Minuten Inaktivität pro Bucket.
- Beim Refill oder bei periodischen Aufräumläufen veraltete Einträge löschen.
- Eine konfigurierbare Obergrenze `maxClients` vorsehen, z. B. 10.000; bei Überschreitung älteste Einträge verwerfen.
- Optional die IP nur als salted Hash speichern, z. B. `sha256(remoteAddr + salt)`, um das Datenschutzrisiko weiter zu senken.
- Die Aufbewahrungsfrist in `SECURITY.md` bzw. `COMPLIANCE.md` dokumentieren.

---

### Befund DSGVO-2 — Unverschlüsselter HTTP-Standardbetrieb möglich (mittel)

**Datei:** `main.go`

Der Server startet standardmäßig über `server.ListenAndServe()` ohne TLS, wenn `TLS_CERT_FILE` und `TLS_KEY_FILE` nicht gesetzt sind. Über diese unverschlüsselte Verbindung würden der Bearer-API-Key und der `user`-Query-Parameter übertragen.

Sofern der Dienst nicht ausschließlich hinter einem TLS-terminierenden Reverse Proxy in einem vertrauenswürdigen Netz betrieben wird, verletzt dies Art. 32 DSGVO.

**Remedy:**

- In `main.go` TLS zum sicheren Standard machen: Ohne `TLS_CERT_FILE`/`TLS_KEY_FILE` nur auf `127.0.0.1:8080` lauschen oder den unverschlüsselten Betrieb nur mit expliziter Umgebungsvariable wie `INSECURE_HTTP=1` erlauben.
- Alternativ in `README.md` und `SECURITY.md` verbindlich dokumentieren, dass der Dienst ausschließlich hinter einem TLS-terminierenden Edge-Proxy exponiert werden darf.
- Der Produktfluss bleibt dabei vollständig funktionsfähig; TLS ändert nur den Transport, nicht die Routen.

---

### Hinweis DSGVO-3 — Rechtsgrundlage für die Verarbeitung der Nutzer-ID dokumentieren (niedrig)

**Datei:** `internal/api/flags_evaluate.go`

Der Parameter `?user=...` ist personenbezogen. Die Implementierung verarbeitet ihn minimal, speichert ihn nicht und die Logging-Middleware protokolliert den Query-String nicht. Die Rechtsgrundlage und eine etwaige Auftragsverarbeitung sind aber organisatorisch beim Betreiber nachzuweisen.

**Empfehlung:**

- In `COMPLIANCE.md` die Rechtsgrundlage für die Verarbeitung der Nutzer-ID dokumentieren.
- Optional bei künftigen Versionen einen `X-User-ID`-Header statt Query-Parameter vorsehen, um das Risiko einer versehentlichen Protokollierung in vorgelagerten Proxy-Logs weiter zu reduzieren.

---

## 2. EU Cyber Resilience Act (CRA)

### Befund CRA-1 — Unbegrenzte Client-Map als Ressourcenerschöpfung (hoch)

**Datei:** `internal/api/ratelimit.go`

Die unbegrenzte `clients`-Map ist auch unter CRA-Gesichtspunkten problematisch. Security by design/default verlangt Schutz vor Ressourcenerschöpfung durch viele unterschiedliche Quell-Adressen.

**Remedy:**

- Wie bei DSGVO-1: TTL, periodische Bereinigung und `maxClients`-Obergrenze einführen.
- Den Mechanismus in `SECURITY.md` als dokumentierte Sicherheitseigenschaft beschreiben.

---

### Positivbefunde CRA

Folgende Maßnahmen sind bereits CRA-konform oder unterstützen die Konformität:

- Body-Limits über `http.MaxBytesReader` und `io.LimitReader`.
- Server-Timeout-Konfiguration in `main.go`.
- TLS-Mindestversion `tls.VersionTLS12`.
- Generische Fehlertexte ohne Stacktraces oder Dateipfade.
- `X-Content-Type-Options: nosniff` und `Cache-Control: no-store`.
- Constant-Time-Vergleich des API-Keys.
- Keine externen Drittanbieter-Abhängigkeiten sichtbar, was die SBOM-Pflicht vereinfacht; `SBOM.md` ist vorhanden.

---

## 3. EU AI Act

Nicht einschlägig. Der Dienst enthält keine KI-Funktion und fällt damit nicht in den Anwendungsbereich des AI Act.

---

## 4. Pflichttexte und UI

Nicht einschlägig. Reines Backend ohne Endnutzer-UI. Es bestehen keine Pflichten zu Impressum, Datenschutzerklärung, Cookie-Banner, AGB oder Widerrufsbelehrung im Code.

---

## 5. Barrierefreiheit

Nicht einschlägig. Keine öffentliche Web-UI vorhanden. WCAG/BITV/EAA-Pflichten greifen nicht.

---

## Fazit

Keine fundamentalen Rechtsverletzungen, die einen sofortigen Stopp erfordern. Zwei behebbare Mängel müssen vor Marktfreigabe behoben werden: die unbefristete IP-Speicherung im Rate-Limiter und der unverschlüsselte HTTP-Standardbetrieb. Danach ist der Dienst aus Sicht der geprüften Bereiche marktreif.