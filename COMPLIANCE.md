VERDICT: CHANGES_REQUESTED

## Beurteilung des zusammengeführten Produktstands

**Projekttyp:** `go-backend` — reines Backend ohne Endnutzer-UI. Daher sind Cookie-/Impressums-/Rechtshinweis- und Barrierefreiheitspflichten für eine öffentliche Web-UI nicht anwendbar. Relevante Pflichten: DSGVO (Datenverarbeitung, Protokollierung, Speicherbegrenzung, Sicherheit), EU Cyber Resilience Act (CRA), EU AI Act nur, falls KI enthalten ist (hier nicht der Fall).

Positiv sind bereits umgesetzt: Request-Größenbegrenzung über `http.MaxBytesReader` (AC-12), generische JSON-Fehlerantworten ohne interne Details (AC-13), kein Logging von `content` oder `id` (AC-14), flüchtige In-Memory-Speicherung ohne Persistenz (AC-15), Thread-Sicherheit über `sync.RWMutex` (AC-10).

---

## 1. DSGVO

### 1.1 Verarbeitete Daten und Rechtsgrundlage
Der Dienst verarbeitet potenziell personenbezogene Daten: Paste-Inhalte (`content`), Sprache (`language`), Zeitstempel (`created_at`, `expires_at`) sowie die technisch erzeugte ID. Die Verarbeitung erfolgt im flüchtigen Prozessspeicher. Eine Rechtsgrundlage nach Art. 6 Abs. 1 lit. b DSGVO (Durchführung des angebotenen Dienstes) ist für den Kernzweck grundsätzlich denkbar. Der Betreiber muss jedoch außerhalb des Backends Transparenz herstellen, da der Code selbst keine Datenschutzerklärung enthält.

### 1.2 Feststellungen

**F1 — Hoch: Fehlende Transportverschlüsselung (Art. 32 DSGVO)**  
`main.go` startet den HTTP-Server ohne TLS: `http.ListenAndServe(":"+port, handler)`. Paste-Inhalte und IDs würden im Klartext über das Netz übertragen. Das ist ein erheblicher Sicherheitsmangel.

**Konkrete Abhilfe:** In `main.go` TLS einbauen oder verpflichtend dokumentieren, dass TLS durch einen vorgelagerten Reverse-Proxy terminiert wird. Beispiel für Direkt-TLS und Timeouts:
```go
srv := &http.Server{
    Addr:              ":" + port,
    Handler:           handler,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       60 * time.Second,
}
if os.Getenv("TLS_CERT") != "" && os.Getenv("TLS_KEY") != "" {
    log.Fatal(srv.ListenAndServeTLS(os.Getenv("TLS_CERT"), os.Getenv("TLS_KEY")))
}
log.Fatal(srv.ListenAndServe())
```
Zusätzlich im `README.md` festhalten, dass Produktivbetrieb nur mit TLS oder TLS-Terminierung erlaubt ist.

**F2 — Hoch: Öffentliche ID-Liste ohne Zugriffskontrolle**  
`internal/api/handlers_read.go` liefert über `GET /pastes` die IDs aller nicht abgelaufenen Pastes. Wer diese Liste abrufen kann, kann anschließend jeden Inhalt über `GET /pastes/{id}` abrufen. Dadurch wird die Vertraulichkeit der Inhalte faktisch aufgehoben. Bei personenbezogenen Inhalten droht eine unbefugte Offenlegung.

**Konkrete Abhilfe:** Die Route `GET /pastes` in `internal/api/handlers_read.go` nur mit Authentifizierung bereitstellen (z. B. konfigurierbares API-Token) oder standardmäßig deaktivierbar machen. Falls eine bewusste öffentliche Pastebin-Funktion gewollt ist, muss dies dokumentiert werden und der Nutzer vor der Veröffentlichung klar informiert werden. Die bloße öffentliche Auflistung aller IDs ohne Schutz ist so nicht konform.

**F3 — Mittel: Unbegrenzte Speicherdauer bei `expires_in_seconds == 0`**  
`internal/paste/store.go` setzt bei `expiresIn == 0` kein Ablaufdatum. Pastes bleiben dauerhaft im Speicher, solange der Prozess läuft. Das widerspricht dem Grundsatz der Speicherbegrenzung (Art. 5 Abs. 1 lit. e DSGVO).

**Konkrete Abhilfe:** In `Store.Create` einen Standard-Ablauf einführen (z. B. 24 Stunden oder 7 Tage), wenn kein Ablauf angegeben wird. Alternativ das Feld `expires_in_seconds` verpflichtend machen. Zusätzlich einen periodischen Cleanup-Job vorsehen, der abgelaufene Einträge aktiv entfernt.

**F4 — Niedrig: Betroffenenrechte nur teilweise abgebildet**  
Ein Löschweg ist vorhanden (`DELETE /pastes/{id}`), aber es gibt keinen dokumentierten Auskunfts- oder Exportweg für Betroffene.

**Konkrete Abhilfe:** Im `README.md` einen Abschnitt „Betroffenenrechte“ aufnehmen: Löschung über `DELETE /pastes/{id}`; Auskunft für den eigenen Paste über `GET /pastes/{id}`; Verweis, dass der Betreiber für weitergehende Anfragen (Berichtigung, Einschränkung, Übertragbarkeit) Prozesse bereitstellen muss.

---

## 2. EU Cyber Resilience Act (CRA)

Das Produkt ist ein „Produkt mit digitalen Elementen“ im Sinne des CRA. Sichtbare Lücken:

**F5 — Hoch: Sicherheit by Design/Default nicht vollständig (Transport und Server-Timeouts)**  
Wie bei F1 fehlt TLS. Zusätzlich werden keine Server-Timeout-Werte gesetzt. Das ermöglicht langsame Verbindungsangriffe (z. B. Slowloris) und entspricht nicht der Anforderung „security by design/default“.

**Konkrete Abhilfe:** In `main.go` einen `http.Server` mit `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` und `IdleTimeout` konfigurieren (siehe F1). TLS entweder direkt aktivieren oder als verbindliche Deployment-Voraussetzung dokumentieren.

**F6 — Mittel: Keine sichtbare SBOM/Sicherheitsdokumentation/Update-Prozess**  
Im sichtbaren Code gibt es keine Dokumentation zu Sicherheitseigenschaften, Abhängigkeiten oder Patch-/Update-Fähigkeit. `go.mod` ist minimal, aber eine explizite SBOM/Sicherheitsdoku fehlt sichtbar.

**Konkrete Abhilfe:** Im `README.md` oder besser in einer neuen `SECURITY.md` folgende Abschnitte ergänzen:
- „Abhängigkeiten/SBOM“: nur Go-Standardbibliothek, keine externen Module; `go.mod` als SBOM-Referenz.
- „Sicherheitseigenschaften“: Body-Limit 1 MiB, generische JSON-Fehlerantworten, In-Memory-Speicherung, zufällige 128-Bit-IDs.
- „Update-/Patch-Prozess“: wie Sicherheitsupdates eingespielt werden.
- „Meldung von Schwachstellen“: Kontaktweg für Sicherheitsmeldungen.

**F7 — Niedrig: Kein Rate-Limiting oder Schutz gegen automatisierte Zugriffe**  
Die zufälligen 128-Bit-IDs sind kryptographisch stark, aber durch die öffentliche Liste (`GET /pastes`) werden sie veröffentlicht. Ein Rate-Limit für die Endpunkte wäre angemessen.

**Konkrete Abhilfe:** Eine einfache Rate-Limiting-Middleware in `internal/api/` vorsehen und in `main.go` einhängen, insbesondere für `POST /pastes`, `GET /pastes`, `GET /pastes/{id}` und `DELETE /pastes/{id}`.

---

## 3. EU AI Act

Kein KI-Feature im Produkt sichtbar. Der EU AI Act ist daher nicht anwendbar.

---

## 4. Pflichttexte und UI

Keine Endnutzer-UI vorhanden. Impressum, Datenschutzerklärung, Cookie-Banner und Widerrufsbelehrung sind für dieses Backend nicht unmittelbar erforderlich. Der Betreiber muss jedoch außerhalb der API eine Datenschutzerklärung bereitstellen. Ein entsprechender Hinweis sollte im `README.md` ergänzt werden.

---

## 5. Barrierefreiheit

Keine öffentliche Web-UI vorhanden. WCAG/BITV/EAA sind für dieses `go-backend` nicht anwendbar.

---

## Reconciliation — Vereinbarkeit mit der Produktfunktion

- **TLS/Timeouts:** Die vorgeschlagenen Änderungen in `main.go` verändern nicht das Handler-Verhalten. Die Tests mit `net/http/httptest` bleiben unverschlüsselt und funktionieren weiterhin. Der Produktionspfad wird lediglich sicherer.
- **`GET /pastes`-Absicherung:** Die Route darf nicht ersatzlos entfernt werden, weil die Spezifikation die Metadatenliste verlangt. Die vorgeschlagene Authentifizierung oder Konfigurationsoption erhält die Funktion für berechtigte Nutzer und schützt zugleich vor unbefugtem Zugriff.
- **Standard-TTL:** Ein Default-Ablauf darf die validen positiven `expires_in_seconds` nicht aushebeln. Positive Werte müssen weiterhin exakt respektiert werden; nur der Fall `0` bzw. fehlender Wert erhält eine Standard-TTL. Damit bleiben die Acceptance Criteria erfüllt.

---

**Gesamtergebnis:** Der Stand ist funktional sauber und erfüllt die Kern-Anforderungen, weist aber behebbare Datenschutz- und CRA-Lücken auf — insbesondere fehlende Transportverschlüsselung, öffentliche ID-Liste, unbegrenzte Speicherdauer und fehlende Sicherheitsdokumentation. Da diese Mängel konkret behebbar sind, lautet das Urteil: `CHANGES_REQUESTED`.