# CLAUDE.md — Backend

> Complementa el [`CLAUDE.md` de la raíz](../../CLAUDE.md). Cómo está armado y cómo se levanta, en [`README.md`](README.md).

## Está congelado

Se terminó en la Fase 1 y **desde ahí no se modifica**. El frontend se escribe contra él, no al revés.

Si algo parece faltar:

1. Casi siempre está y hay que buscarlo mejor — `openapi.yaml` tiene los 13 endpoints.
2. Si de verdad falta, se agrega **primero en [`docs/contract/`](../../docs/contract/CLAUDE.md)** y se conversa. **Preguntá antes de escribir.**
3. Nunca se cambia una respuesta del backend para acomodar al frontend. Si la forma molesta, el que está mal puede ser el contrato — y eso también se conversa.

## Lo único que se toca sin preguntar

Un bug real: algo que no cumple lo que dice `docs/contract/`. En ese caso el arreglo va con su test, y `go test ./...` tiene que quedar verde.

## Verificar

```bash
go test ./...              # incluye los golden contra docs/contract/samples/
gofmt -l . && go vet ./...
```

Los dos que importan:

- **`TestSimulateReproducesFixture`** — Go y JavaScript producen la misma carrera, tick por tick. Es lo que hace que la app se vea igual contra el mock y contra el servidor.
- **`TestRespuestasCoincidenConLosSamples`** — la forma de cada respuesta coincide con el contrato. Compara **nombres de campo**, no valores: los valores cambian con el rebase de fechas, los nombres son lo que rompe al frontend en silencio.
