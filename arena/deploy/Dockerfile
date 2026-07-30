# Imagen del backend de Arena. Dos etapas: se compila y se copia el binario a una
# imagen sin shell ni gestor de paquetes.
#
# El contexto de build es `arena/`, no `arena/backend/`, para que el Dockerfile
# pueda vivir junto al resto del despliegue. Desde la raíz del repo:
#
#   docker build -f arena/deploy/Dockerfile -t arena-api arena
#   docker run --rm -p 127.0.0.1:8080:8080 --env-file arena/deploy/.env arena-api
#
# El puerto se publica **solo en loopback**. Nada de este despliegue abre un
# puerto de entrada: `cloudflared` sale hacia Cloudflare desde adentro del VPS.

FROM golang:1.26-alpine AS build

WORKDIR /src

# Las dependencias primero: mientras go.mod y go.sum no cambien, Docker reusa
# esta capa y no vuelve a descargar nada.
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

# CGO_ENABLED=0 da un binario estático, que es lo que permite la imagen final sin
# libc. -trimpath saca las rutas de compilación y -s -w la tabla de símbolos.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/arena .

# distroless/static trae los certificados raíz. Hoy la base es Postgres en el
# mismo VPS y la conexión no sale del bridge de Docker, pero el día que esté
# afuera (`sslmode=require`) el TLS funciona sin tocar la imagen. No trae shell:
# si alguien entra al contenedor, no tiene con qué moverse.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/arena /arena

# El esquema viaja en la imagen para poder aplicarlo desde el mismo artefacto que
# se despliega, sin clonar el repo en el VPS.
COPY docs/contract/schema.sql /contract/schema.sql

ENV PORT=8080
EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/arena"]
