package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"slices"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Los .sql se aplican en orden lexicográfico. `0001_contract_schema.sql` es una
// copia literal de docs/contract/schema.sql —el contrato manda— y lo verifica
// TestElEsquemaEmbebidoCoincideConElContrato.
//
//go:embed migrations/*.sql
var migrationFS embed.FS

const migrationsDir = "migrations"

// advisoryLock es la llave del lock de aplicación. Un número arbitrario y fijo:
// lo único que importa es que los dos procesos que arranquen a la vez elijan el
// mismo. Sin el lock, dos instancias corriendo `create type` simultáneamente se
// pisan y una de las dos muere en el arranque.
const advisoryLock = 0x4152454e41 // "ARENA" en hexa

// Migrate aplica el esquema. **Es idempotente**: correrlo dos veces no falla.
//
// No hay tabla de versiones y no hace falta: cada sentencia del esquema está
// escrita para poder repetirse —`create table if not exists`, `create or replace
// function`, y `create type` envuelto en un bloque que ignora duplicate_object—.
// El esquema se vuelve a aplicar en cada arranque y eso es una feature: el
// despliegue no tiene un paso aparte que alguien pueda olvidarse de correr.
func Migrate(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	files, err := migrationFiles()
	if err != nil {
		return err
	}

	// Todo en la MISMA conexión: un advisory lock de sesión vive en la conexión
	// que lo tomó, así que soltar la conexión sería soltar el lock.
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("tomando una conexión para migrar: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "select pg_advisory_lock($1)", int64(advisoryLock)); err != nil {
		return fmt.Errorf("tomando el lock de migración: %w", err)
	}
	defer func() {
		if _, err := conn.Exec(ctx, "select pg_advisory_unlock($1)", int64(advisoryLock)); err != nil {
			log.Error("no se pudo soltar el lock de migración", "error", err)
		}
	}()

	for _, name := range files {
		content, err := fs.ReadFile(migrationFS, migrationsDir+"/"+name)
		if err != nil {
			return fmt.Errorf("leyendo %s: %w", name, err)
		}

		// Exec SIN argumentos manda el archivo entero por el protocolo simple, que
		// es el único que acepta varias sentencias en un envío. Con argumentos, pgx
		// usaría el protocolo extendido y Postgres rechazaría el archivo completo.
		// Postgres lo corre en una transacción implícita: si una sentencia falla,
		// no queda medio esquema aplicado.
		if _, err := conn.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("aplicando %s: %w", name, err)
		}
		log.Debug("migración aplicada", "archivo", name)
	}

	log.Info("esquema al día", "migraciones", len(files))
	return nil
}

func migrationFiles() ([]string, error) {
	entries, err := fs.ReadDir(migrationFS, migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("leyendo las migraciones embebidas: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	slices.Sort(names)
	return names, nil
}
