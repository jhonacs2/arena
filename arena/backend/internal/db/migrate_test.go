package db

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"testing"
	"time"
)

// El contrato manda (docs/contract/CLAUDE.md): `0001_contract_schema.sql` es una
// COPIA de docs/contract/schema.sql, no una versión del backend.
//
// Se copia y no se lee del disco en producción porque `go:embed` no puede salir
// del directorio del paquete y un binario que necesita archivos al lado no es un
// binario que se despliegue tranquilo. El precio de la copia es que puede quedar
// desfasada, y este test es lo que hace que no pase en silencio.
const contractSchemaPath = "../../../docs/contract/schema.sql"

func TestElEsquemaEmbebidoCoincideConElContrato(t *testing.T) {
	want, err := os.ReadFile(contractSchemaPath)
	if err != nil {
		t.Fatalf("leyendo el esquema del contrato: %v", err)
	}
	got, err := fs.ReadFile(migrationFS, migrationsDir+"/0001_contract_schema.sql")
	if err != nil {
		t.Fatalf("leyendo el esquema embebido: %v", err)
	}

	if !bytes.Equal(normalize(want), normalize(got)) {
		t.Fatalf("el esquema embebido quedó desfasado del contrato.\n" +
			"Sincronizalo con:\n" +
			"  cp arena/docs/contract/schema.sql arena/backend/internal/db/migrations/0001_contract_schema.sql")
	}
}

// normalize iguala los finales de línea: el repo se trabaja en Windows y un
// checkout con CRLF no debería hacer fallar el test.
func normalize(content []byte) []byte {
	return bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
}

func TestLasMigracionesEstanOrdenadas(t *testing.T) {
	files, err := migrationFiles()
	if err != nil {
		t.Fatalf("migrationFiles: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no hay migraciones embebidas")
	}
	if files[0] != "0001_contract_schema.sql" {
		t.Errorf("la primera migración es %q; el esquema del contrato tiene que ir primero", files[0])
	}
	for i := 1; i < len(files); i++ {
		if files[i-1] >= files[i] {
			t.Errorf("las migraciones no están ordenadas: %q antes de %q", files[i-1], files[i])
		}
	}
}

// El requisito de arena/CLAUDE.md §7: «que el esquema aplique limpio dos veces
// seguidas». Es lo que permite aplicarlo en cada arranque sin un paso manual que
// alguien se pueda olvidar.
func TestMigrarDosVecesNoFalla(t *testing.T) {
	url := os.Getenv("ARENA_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("sin ARENA_TEST_DATABASE_URL: se saltea el test de base")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := Open(ctx, url, Options{MaxConns: 4})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer pool.Close()

	// El mismo lock de turno que usa internal/testdb, repetido acá a mano: este
	// paquete no puede importar testdb —testdb importa db y sería un ciclo—. Sin
	// el turno, migrar mientras otro paquete trunca las mismas tablas es un
	// deadlock intermitente.
	const exclusiveLock = 0x4152454e41544553 // "ARENATES"
	lockConn, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("tomando la conexión del lock: %v", err)
	}
	defer lockConn.Release()
	if _, err := lockConn.Exec(context.Background(), `select pg_advisory_lock($1)`, int64(exclusiveLock)); err != nil {
		t.Fatalf("tomando el lock: %v", err)
	}
	defer func() {
		if _, err := lockConn.Exec(context.Background(), `select pg_advisory_unlock($1)`, int64(exclusiveLock)); err != nil {
			t.Errorf("soltando el lock: %v", err)
		}
	}()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	for intento := 1; intento <= 2; intento++ {
		if err := Migrate(ctx, pool, log); err != nil {
			t.Fatalf("aplicación %d: %v", intento, err)
		}
	}

	// Y la vista quedó usable: aplicar dos veces un `create or replace view` es
	// donde se rompería si el orden de las columnas cambiara.
	var count int
	if err := pool.QueryRow(ctx, `select count(*) from user_scores`).Scan(&count); err != nil {
		t.Fatalf("user_scores no se puede consultar después de migrar dos veces: %v", err)
	}
}
