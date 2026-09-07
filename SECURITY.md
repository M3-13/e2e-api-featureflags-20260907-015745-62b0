VERDICT: CHANGES_REQUESTED

Der Scanner-Output enthält für dieses Projekt (`go-backend`) keinen anwendbaren Security-Scanner. Ich habe daher den vollständigen Produktstand manuell geprüft. Es wurden keine hartkodierten Secrets, keine klassische Injection/RCE, kein Auth-Bypass und keine PII-Leaks im Antwort-Body oder Logging gefunden. Es bestehen jedoch zwei mittlere Härtungsdefizite sowie kleinere Punkte, die vor einem öffentlichen Deployment behoben werden sollten.

## Sicherheitsbericht

### 1. [Mittel] Fehlende TLS-Pflicht für den API-Key-Transport
**Betroffene Stelle:** `main.go`, insbesondere der Fall `server.ListenAndServe()` ohne TLS.

**Befund:**  
Der Dienst erwartet einen Bearer-API-Key (`Authorization: Bearer ...`). Wenn `TLS_CERT_FILE`/`TLS_KEY_FILE` nicht gesetzt sind, startet der Server unverschlüsselt auf `:8080`. Damit wird der API-Key bei jedem Request im Klartext übertragen und kann bei einem Netzwerk-Mitschnitt abfließen. Das ist insbesondere dann relevant, wenn der Dienst ohne vorgeschaltetes TLS-Terminierungs-Proxy betrieben wird.

**Konkrete Härtung:**  
TLS nicht als optionalen Sonderfall behandeln, sondern als sicheren Standard erzwingen:

```go
if certFile == "" || keyFile == "" {
    log.Fatal("TLS_CERT_FILE und TLS_KEY_FILE müssen gesetzt sein")
}
```

Falls ausdrücklich unverschlüsselter Betrieb für lokale Entwicklung erlaubt sein soll, diesen nur über eine explizite Opt-in-Variable wie `ALLOW_INSECURE_HTTP=true` starten und deutlich warnen. Die bestehenden TLS-Pfade (konfigurierte Zertifikate) bleiben dabei voll funktionsfähig.

---

### 2. [Mittel] Rate-Limit-Client-Map wächst unbegrenzt
**Betroffene Stelle:** `internal/api/ratelimit.go`

**Befund:**  
Die Middleware speichert pro `r.RemoteAddr` einen `bucket` in der Map `clients`. Es gibt kein Entfernen veralteter Einträge. Bei vielen unterschiedlichen Client-Adressen wächst die Map dauerhaft und hält Speicher – ein langsam wirkender Speicher-DoS. Da `RemoteAddr` bei direkten Zugriffen aus der TCP-Verbindung stammt, ist das nicht beliebig fälschbar, aber das Risiko besteht bei hoher Client-Vielfalt oder über längere Zeit.

**Konkrete Härtung:**  
Periodisch veraltete Buckets entfernen, z. B. während eines Requests oder über einen Hintergrund-Ticker:

```go
// nach dem Lock, bei jedem Request oder zeitgesteuert:
for addr, b := range clients {
    if now.Sub(b.last) > 5*time.Minute {
        delete(clients, addr)
    }
}
```

Alternativ eine maximale Anzahl Clients einführen und bei Überschreitung die ältesten Einträge verwerfen. Die eigentliche Rate-Limit-Funktion bleibt unverändert.

---

### 3. [Niedrig] `RemoteAddr` als Rate-Limit-Identität hinter Proxys
**Betroffene Stelle:** `internal/api/ratelimit.go`, Zeile mit `r.RemoteAddr`.

**Befund:**  
Hinter einem Reverse-Proxy sehen alle Requests für den Backend-Prozess nach derselben Quell-IP aus. Dann teilen sich alle Nutzer denselben Token-Bucket; ein einzelner Angreifer kann das Limit für alle erschöpfen (geteilter Denial-of-Service). Das ist keine direkte Schwachstelle im Code, aber deploymentsensibel.

**Konkrete Härtung:**  
`X-Forwarded-For` nur auswerten, wenn der Dienst direkt hinter einem vertrauenswürdigen Proxy läuft, und dabei nur die letzte vom Proxy ergänzte IP verwenden. Alternativ dokumentieren, dass diese Middleware nicht für Proxy-Setups mit gemeinsamer Quell-IP geeignet ist.

---

### 4. [Niedrig] Standard-ServeMux-Fehler sind kein JSON-Fehlerobjekt
**Betroffene Stelle:** `main.go` / `http.NewServeMux()`, innere Mux-Routen.

**Befund:**  
Unbekannte Pfade oder nicht erlaubte HTTP-Methoden in der inneren Mux werden vom Standard-`http.ServeMux` als Textantworten wie `404 page not found` bzw. `405 Method Not Allowed` beantwortet. Sie enthalten keine internen Details und sind daher kein direktes Sicherheitsleck, verletzen aber die produktspezifische Anforderung an generische JSON-Fehlerobjekte.

**Konkrete Härtung:**  
Für die innere Mux einen eigenen Not-Found-/Method-Not-Allowed-Fallback registrieren, der `api.WriteError(w, http.StatusNotFound, "not found")` bzw. `api.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")` nutzt. Dabei weiterhin sicherstellen, dass keine internen Fehlertexte exponiert werden.

---

## Positiv geprüft

- Keine hartkodierten Schlüssel oder Passwörter; `API_KEY` kommt ausschließlich aus der Umgebung.
- Zugriffs-Logging protokolliert nur Methode, Pfad ohne Query-String und Statuscode; der `user`-Parameter wird nicht geloggt.
- SQL/Command-Deserialisierung nicht vorhanden; JSON-Decoding erfolgt mit Größenbegrenzung (`MaxBytesReader` bzw. `LimitReader`) und liefert 413 bei Überschreitung.
- Key-Validierung (`^[A-Za-z0-9._-]{1,128}$`) verhindert unerwartete Zeichen; der Schlüssel wird nur als Map-Key verwendet.
- Rollout-Hash (`FNV-64a`) ist deterministisch und für die Feature-Entscheidung unkritisch.
- Fehlerantworten der API-Handler sind generisch und enthalten keine Stacktraces oder Dateipfade.
- `RequireAPIKey` nutzt `subtle.ConstantTimeCompare` für den API-Key-Vergleich.
- Antworten setzen `X-Content-Type-Options: nosniff` und `Cache-Control: no-store`.
- Store-Zugriffe sind durch `sync.RWMutex` gegen Race Conditions geschützt.
- TLS-Konfiguration erzwingt mindestens TLS 1.2, sofern TLS aktiv ist.

## Fazit

Das Produkt ist für eine interne oder durch TLS-Terminierung geschützte Umgebung grundsätzlich solide. Vor einem ungeschützten öffentlichen Betrieb sollten die beiden mittleren Punkte (TLS-Pflicht und Rate-Limit-Speicherbereinigung) behoben werden. Da keine hochkritische oder kritische Schwachstelle erkennbar ist, lautet das Gesamturteil `CHANGES_REQUESTED`.