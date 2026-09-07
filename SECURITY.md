VERDICT: BLOCKED

## Sicherheitsbericht

### Zusammenfassung der Prüfung
Es wurde eine manuelle Sicherheitsanalyse des Go-Backends durchgeführt. Ein automatisierter Scanner war laut Angabe nicht anwendbar; daher beruhen die folgenden Bewertungen ausschließlich auf der Code-Analyse.

Die Umsetzung der Acceptence-Kriterien AC-11 bis AC-14 (RequestBody-Limits, generische JSON-Fehler, Logging ohne Query-String) ist im Code ordnungsgemäß umgesetzt.  
Dennoch besteht eine **kritische** Lücke im Bereich Authentifizierung und Autorisierung, die einen sicheren Produktivbetrieb ohne zusätzliche Infrastruktur nicht zulässt.

### Prüfbereiche im Detail

#### 1. Secrets
Keine hartkodierten Passwörter, API-Keys, Token oder sonstige Secrets im Code gefunden.  
Keine Secrets in Logs erkennbar.

#### 2. Injection & Eingaben
- **SQL/Command Injection:** Nicht anwendbar – keine Datenbank, keine Shell-Aufrufe.
- **Path Injection:** Die Pfadvariable `key` wird nicht für Dateisystem- oder Pfadoperationen genutzt; sie fließt lediglich in eine In-Memory-Map und in einen Hash ein. Keine direkte Path-Injection.
- **JSON-Deserialisierung:** Nutzung der sicheren Go-Standardbibliothek; kein erkennbares Unsafe-Deserialization-Risiko.
- **RequestBody-Limits:** Für `POST /flags` wird `http.MaxBytesReader` verwendet, für `PUT /flags/{key}` ein `io.LimitReader` mit anschließender Längenprüfung. Bei Überschreitung erfolgt Status 413 mit JSON-Fehlerobjekt.
- **Fehlerantworten:** Alle sichtbaren Fehlerantworten sind generisch (`{"error": ...}`). Interne Fehlermeldungen, Stacktraces oder Dateipfade werden nicht exponiert.
- **Logging:** Die Middleware protokolliert ausschließlich Methode, URL-Pfad ohne Query-String und Statuscode. Der `user`-Parameter aus `/flags/{key}/evaluate?user=...` wird nicht geloggt.

Schwachstelle in diesem Bereich:  
- **Unzureichende Key-Validierung** (siehe Finding F003).

#### 3. AuthN/AuthZ
- **Keine Authentifizierung oder Autorisierung** für irgendeinen Endpunkt vorhanden.  
- Jeder Client, der den HTTP-Port erreicht, kann Feature-Flags anlegen, ändern, löschen und deren Rollout-Prozent beeinflussen.  
- Der Server bindet mit `:8080` an **alle** Netzwerkinterfaces (`0.0.0.0`), nicht nur an ein vertrauenswürdiges internes Interface wie `127.0.0.1`.

Dies wird als **kritisch** eingestuft (Finding F001).

#### 4. Abhängigkeiten
- Es werden ausschließlich Pakete der Go-Standardbibliothek verwendet (`net/http`, `encoding/json`, `sync`, `hash/fnv`, `log`).
- `go.mod` enthält gemäß Umfang keine externen Third-Party-Dependencies.
- Keine bekannten CVEs oder veralteten Pakete sichtbar.

#### 5. Konfiguration & Transport
- Der Dienst läuft als reiner HTTP-Server ohne TLS.
- Feature-Flag-Konfiguration sowie der `user`-Query-Parameter werden im Klartext übertragen.
- Keine Sicherheitsheader gesetzt (für eine reine JSON-API weniger kritisch, aber dennoch zu beachten).
- Kein Rate-Limiting oder sonstiger DoS-Schutz (siehe Finding F004).

---

### Findings

#### F001 – Fehlende Authentifizierung und Autorisierung (Kritisch / Hoch)
**Betroffen:**  
`main.go` (Routing ohne Auth-Middleware), `internal/api/flags_create.go`, `internal/api/flags_update.go`, `internal/api/flags_delete.go`

**Beschreibung:**  
Der Feature-Flag-Service stellt ungeschützte Schreibendpunkte bereit. Ein Angreifer, der den Port erreicht, kann:
- beliebige Flags anlegen (`POST /flags`),
- bestehende Flags verändern (`PUT /flags/{key}`),
- Flags löschen (`DELETE /flags/{key}`),
- die Rollout-Logik und damit das Verhalten der abhängigen Anwendung manipulieren.

Da der Server auf `:8080` lauscht, ist er potenziell von außen erreichbar. Eine vorgelagerte Authentifizierung durch ein API-Gateway oder eine private Netzwerksegmentierung ist im Code nicht sichtbar und kann daher nicht vorausgesetzt werden.

**Konkreter Fix:**  
- Einführung einer Authentifizierungs-Middleware für **alle** Routen, z. B. statischer Bearer-Token aus einer Umgebungsvariable.  
- Alternativ: Bind an ein privates Interface (`127.0.0.1` oder internes Pod-Netzwerk) und erzwinge Authentifizierung am vorgelagerten Gateway.  
- Die direkten Handler-Tests bleiben unverändert lauffähig, wenn die Middleware nur in `main.go` beim Zusammenbau des `http.Handler` eingefügt wird.

**Severity:** Hoch

---

#### F002 – Unverschlüsselte HTTP-Übertragung (Mittel)
**Betroffen:**  
`main.go` (`http.Server{Addr: ":8080", Handler: handler}`)

**Beschreibung:**  
Die gesamte Kommunikation erfolgt über unverschlüsseltes HTTP. Alle übermittelten Daten, einschließlich der `user`-IDs bei der Evaluierung und der Feature-Flag-Konfiguration, sind für Angreifer im Netzwerk einsehbar.

**Konkreter Fix:**  
- TLS aktivieren, z. B. durch `server.ListenAndServeTLS(certFile, keyFile)` oder durch einen vorgelagerten TLS-terminierenden Reverse-Proxy.  
- Alternativ, falls der Dienst ausschließlich in einem isolierten internen Netz ohne sensible Daten betrieben wird, die Bindung auf ein privates Interface beschränken und diese Entscheidung dokumentieren.

**Severity:** Mittel

---

#### F003 – Unzureichende Validierung des Feature-Flag-Keys (Niedrig)
**Betroffen:**  
`internal/api/flags_create.go` (keine Längen-/Zeichenprüfung), `internal/api/flags_evaluate.go` (Hash-Eingabe), `internal/store/store.go` (Map-Key)

**Beschreibung:**  
Bei `POST /flags` wird nur geprüft, ob `key` nicht leer ist. Es gibt keine Begrenzung der Länge und keine Einschränkung der erlaubten Zeichen. Ein Angreifer (sofern Zugriff besteht) könnte sehr lange oder Sonderzeichen-enthaltende Keys anlegen. Dies kann zu erhöhtem Speicherverbrauch führen und die Verwendung solcher Keys in URL-Pfaden erschweren oder zu unerwartetem Routing-Verhalten führen.

**Konkreter Fix:**  
- Key-Format validieren, z. B. per Regex `^[A-Za-z0-9._-]{1,128}$`.  
- Diese Einschränkung ist mit der aktuellen Spec kompatibel, da keine anderen Key-Formate gefordert sind.  
- Bei Verstoß Status 400 mit `{"error":"invalid key format"}` zurückgeben.

**Severity:** Niedrig

---

#### F004 – Fehlendes Rate-Limiting bzw. Schutz vor Überlastung (Niedrig)
**Betroffen:**  
`main.go`, `internal/api/middleware.go`

**Beschreibung:**  
Es gibt zwar Body-Limits für Schreiboperationen, aber keine Begrenzung der Anfragehäufigkeit. Ein erreichbarer Dienst kann durch massenhafte Anfragen (z. B. viele `POST /flags`) Speicher und CPU belasten. Dies ist insbesondere in Kombination mit F001 relevant.

**Konkreter Fix:**  
- Rate-Limiting-Middleware einführen (z. B. Token Bucket pro Client-IP oder API-Key), sofern der Dienst nicht ausschließlich in einem stark abgeschotteten Netzwerk betrieben wird.  
- Alternativ die Schreibendpunkte nur über ein vorgelagertes Gateway mit eigenem Rate-Limiting erreichbar machen.

**Severity:** Niedrig

---

### Fazit
Der Code erfüllt die spezifizierten Sicherheitsanforderungen bezüglich Body-Limits, generischer Fehlerantworten und datenschutzfreundlichem Logging.  
Die **fehlende Authentifizierung und Autorisierung** an den schreibenden Endpunkten stellt jedoch ein erhebliches Risiko dar, das einen sicheren Betrieb ohne zusätzliche Schutzmaßnahmen nicht zulässt. Daher wird das Produkt als **BLOCKED** eingestuft.

Behebung: Mindestens Finding F001 muss umgesetzt werden, bevor der Dienst für den Kunden freigegeben werden kann. F002 sollte im selben Zuge adressiert werden.