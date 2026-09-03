VERDICT: CHANGES_REQUESTED

## Sicherheitsreport

### Prüfbereiche im Überblick

- **Secrets:** Keine hartkodierten Schlüssel, Passwörter oder Tokens sichtbar. Die einzigen Logausgaben (`pastebin API listening on port …`, `server failed: …`) protokollieren weder `content` noch Paste-`id`. AC-14 ist erfüllt.
- **Injection & Eingaben:** POST begrenzt den Body vor dem Einlesen über `http.MaxBytesReader` auf 1 MB. JSON wird typsicher dekodiert. Fehlerantworten sind generisch und ohne interne Details (AC-13). Pfad-IDs werden ausschließlich als Map-Schlüssel verwendet; keine SQL-, Command- oder Pfad-Injektion erkennbar.
- **AuthN/AuthZ:** Der Service ist bewusst ohne Authentifizierung. Zugriff erfolgt über zufällige, kryptografisch erzeugte 128-Bit-IDs. GET `/pastes` gibt gemäß AC-07 niemals `content` aus.
- **Abhängigkeiten:** Das Projekt nutzt nur die Go-Standardbibliothek. Es wurden keine Scanner-Ergebnisse für diesen Projekttyp geliefert; es sind keine anfälligen externen Abhängigkeiten sichtbar.
- **Konfiguration & Transport:** Grundlegende Konfiguration ist unkritisch. Es fehlen jedoch Server-Timeouts, eine Speicher-/Ratenbegrenzung und eine TLS-Terminierung.

### Findings

#### 1. Unbegrenztes Speicherwachstum durch fehlende Bereinigung abgelaufener Pastes

- **Schweregrad:** mittel
- **Betroffene Stelle:** `internal/paste/store.go`, insbesondere `Create`, `Get` und `List`
- **Problem:** Abgelaufene Pastes werden von `Get` und `List` nur ausgeblendet, aber nie aus der internen Map gelöscht. Ein anonymer Client kann fortlaufend Pastes mit bis zu 1 MB Inhalt erstellen. Ohne Gesamtlimit wächst der Prozessspeicher unbegrenzt, was zu einem Denial of Service führt.
- **Konkrete Lösung:** Beim `Create` opportunistisch abgelaufene Einträge entfernen:
  ```go
  s.mu.Lock()
  for id, p := range s.pastes {
      if s.IsExpired(p) {
          delete(s.pastes, id)
      }
  }
  s.pastes[id] = p
  s.mu.Unlock()
  ```
  Zusätzlich eine Obergrenze für die Gesamtzahl oder Gesamtbytes einführen. Zum Beispiel kann der Handler nach dem `Create` prüfen, ob ein Schwellwert überschritten wurde, und im Fehlerfall mit HTTP 503 antworten — ohne die eigene Produktfunktion zu brechen.

#### 2. Fehlende HTTP-Server-Timeouts

- **Schweregrad:** mittel
- **Betroffene Stelle:** `main.go`, `http.ListenAndServe(":"+port, handler)`
- **Problem:** Der Server wird mit den Standard-Timeouts betrieben. Langsame Clients können Verbindungen unnötig lange offen halten (Slowloris-ähnlicher Ressourcenverbrauch).
- **Konkrete Lösung:** Einen expliziten `http.Server` mit Timeouts verwenden:
  ```go
  srv := &http.Server{
      Addr:              ":" + port,
      Handler:           handler,
      ReadHeaderTimeout: 5 * time.Second,
      ReadTimeout:       10 * time.Second,
      WriteTimeout:      20 * time.Second,
      IdleTimeout:       60 * time.Second,
  }
  log.Printf("pastebin API listening on port %s", port)
  if err := srv.ListenAndServe(); err != nil {
      log.Fatalf("server failed: %v", err)
  }
  ```
  Zusätzlich den Import `time` ergänzen. Die gewählten Timeouts erlauben weiterhin 1-MB-Uploads und normale Antwortzeiten.

#### 3. Keine Transportverschlüsselung bei direkter Exposition

- **Schweregrad:** mittel
- **Betroffene Stelle:** `main.go`
- **Problem:** Der Dienst lauscht auf reinem HTTP. Paste-Inhalte und IDs werden im Klartext übertragen. Sofern kein TLS-terminierender Reverse-Proxy vorgeschaltet ist, können Dritte im Netzwerk Inhalte mitlesen.
- **Konkrete Lösung:** Entweder TLS direkt im Server aktivieren, z. B. `http.ListenAndServeTLS` mit Zertifikats- und Schlüsseldateien aus Umgebungsvariablen, oder in der Betriebsdokumentation verbindlich festlegen, dass der Dienst ausschließlich hinter einem TLS-terminierenden Proxy erreichbar sein darf. Beides schränkt die vorhandenen Produktfunktionen nicht ein.

### Bewertung

Es wurden keine kritischen oder hohen Schwachstellen gefunden. Die Kernanforderungen der Sicherheits-Akzeptanzkriterien (AC-12, AC-13, AC-14, AC-15) sind im Code sichtbar umgesetzt. Die drei mittleren Härtungsbedarfe betreffen Ressourcenerschöpfung, Verbindungs-DoS und Transportvertraulichkeit. Daher ist das Produkt nicht abzulehnen, sollte aber vor dem Produktivgang nachgebessert werden.