# Imagen del backend. Dos etapas: se compila y se copia el binario a una
# imagen sin shell ni gestor de paquetes.
#
#   docker build -t hipodromo-api project/backend
#   docker run -p 8080:8080 -e JWT_SECRET=… hipodromo-api

FROM golang:1.26-alpine AS build

WORKDIR /src

# Las dependencias se copian primero: mientras go.mod y go.sum no cambien,
# Docker reusa esta capa y no vuelve a descargar nada.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 da un binario estático que corre en una imagen sin libc.
# -trimpath saca las rutas de compilación; -s -w saca la tabla de símbolos.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/hipodromo .

# El dataset va embebido en el binario, así que la imagen final no lleva
# archivos sueltos: solo el ejecutable y los certificados raíz que necesita
# la llamada HTTPS a Resend.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/hipodromo /hipodromo

# La copia en disco necesita un directorio escribible. Sin volumen montado,
# el servidor avisa una vez y sigue funcionando solo en memoria.
VOLUME /data
ENV SNAPSHOT_PATH=/data/snapshot.json

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/hipodromo"]
