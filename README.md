# Pastebin-REST-API in Go

Eine kleine, thread-sichere Pastebin-REST-API in Go, die ausschließlich die
Standardbibliothek (`net/http`) nutzt. Pastes werden in einem flüchtigen
In-Memory-Store mit Mutex gehalten, können optional ablaufen und sind über
`POST`/`GET`/`DELETE` erreichbar; `GET /pastes` liefert Metadaten ohne Inhalt.

## Tech Stack

- **Sprache**: Go (nur Standardbibliothek, keine Drittmodule)
- **HTTP**: `net/http` (Standardbibliothek)
- **Speicherung**: In-Memory mit `sync.RWMutex`
- **IDs**: `crypto/rand`, hex-kodiert (32 Zeichen)
- **Tests**: `go test` mit `net/http/httptest`

## Installation / Voraussetzungen

Go 1.22 oder neuer. Keine weiteren Abhängigkeiten.

## Starten (Dev)

```sh
go run .
```

Der Server lauscht standardmäßig auf Port **8080**. Über die Umgebungsvariable
`PORT` lässt sich ein anderer Port wählen:

```sh
PORT=9000 go run .
```

(Wenn `PORT` nicht gesetzt ist, wird 8080 verwendet.)

## Tests

```sh
go test ./...
```

## Endpunkt-Übersicht

Fehlerantworten liefern immer `{"error":"..."}`, Erfolgsantworten
`application/json`.

| Methode | Pfad           | Beschreibung                                                        |
|---------|----------------|---------------------------------------------------------------------|
| GET     | `/health`      | Liveness: `200` mit `{"status":"ok"}`                                |
| POST    | `/pastes`      | Legt einen Paste an; Body `{"content":str,"language":str?,"expires_in_seconds":int>0?}` → `201` mit `{"id":"<32-hex>"}` |
| GET     | `/pastes/{id}` | Liefert den vollständigen Paste inkl. `content` → `200`; unbekannt/abgelaufen → `404` |
| GET     | `/pastes`      | Liefert Metadaten aller nicht abgelaufenen Pastes (nie `content`)    |
| DELETE  | `/pastes/{id}` | Löscht den Paste → `204`; unbekannt → `404`                          |

Unbekannter Pfad → `404`, falsche Methode auf bekanntem Pfad → `405` mit
`Allow`-Header.

## Features

- In-Memory-Store mit Mutex (race-sicher, kein Datenrennen)
- Optionaler Ablauf von Pastes über `expires_in_seconds`
- Zufällige 32-Zeichen-IDs aus `crypto/rand`
- Keine persistente Ablage – Pastes leben nur im Prozessspeicher
- Logausgaben enthalten ausschließlich technische Meldungen (nie `content` oder `id`)
