// Package testdb levanta una base de prueba para los tests que necesitan
// Postgres.
//
// Las reglas que importan de Arena —el canje atómico de un código, que el ledger
// y `users.balance` no se separen, el piso del saldo— **no se pueden probar sin
// una base de verdad**. Son reglas que hacen cumplir un lock de fila, un CHECK y
// un trigger; un mock en memoria probaría el mock.
//
// Cómo correrlos:
//
//	docker run -d --rm --name arena-pg-test -e POSTGRES_PASSWORD=test \
//	  -e POSTGRES_DB=arena_test -p 55433:5432 postgres:16-alpine
//
//	ARENA_TEST_DATABASE_URL="postgres://postgres:test@localhost:55433/arena_test?sslmode=disable" \
//	  go test ./...
//
// **La base tiene que llamarse con «test» en el nombre**, y no es una convención:
// `Pool` corta si no. Los tests truncan todas las tablas, y apuntar esta variable
// a la base de desarrollo ya se llevó puestos los datos de alguien una vez.
//
// Sin la variable, los tests de base **se saltean** en vez de fallar: `go test
// ./...` tiene que poder correr en una máquina sin Docker. Los tests de lógica
// pura (auth, invite, puntos) corren siempre.
package testdb

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/talentodh/arena/internal/db"
)

// EnvVar es la variable con la cadena de conexión de prueba.
//
// Es una variable DISTINTA de DATABASE_URL a propósito: los tests borran todas
// las tablas, y un `go test` corrido con la DATABASE_URL de producción en el
// entorno sería la peor forma posible de aprender esta lección.
const EnvVar = "ARENA_TEST_DATABASE_URL"

// exclusiveLock serializa los tests que usan la base.
//
// Hace falta porque `go test ./...` corre los paquetes EN PARALELO y todos
// apuntan a la misma base: mientras un test de `ledger` mueve monedas, el primer
// test de `api` hace TRUNCATE y le borra el escenario abajo. Pasa de verdad, y se
// ve como un `USER_NOT_FOUND` intermitente o como un deadlock —el TRUNCATE pide
// un lock exclusivo de tabla mientras el otro tiene locks de fila tomados—.
//
// Un lock de sesión de Postgres lo resuelve para todos los paquetes a la vez, sin
// tener que acordarse de pasar `-p 1`. Un test que espera es preferible a un test
// que falla una vez cada cinco.
const exclusiveLock = 0x4152454e41544553 // "ARENATES" en hexa

// refuseNonTestDatabase corta si la base no se llama como una base de prueba.
//
// **Esto ya pasó, y por eso existe.** Alguien apuntó ARENA_TEST_DATABASE_URL a la
// base del `docker-compose` de desarrollo para correr los tests rápido. Los tests
// truncan todas las tablas: se llevaron puesta la carrera y el alumno que había
// cargados a mano para probar la app. Nada falló, nada avisó; los tests dieron
// verde, que es lo peor que podía pasar.
//
// Una variable distinta no alcanzaba como protección: distingue el descuido de
// exportar DATABASE_URL, pero no el de escribir a mano la misma cadena. El nombre
// de la base sí, porque es una decisión explícita al crear el contenedor.
func refuseNonTestDatabase(url string) error {
	name := url
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if i := strings.IndexAny(name, "?"); i >= 0 {
		name = name[:i]
	}
	if strings.Contains(strings.ToLower(name), "test") {
		return nil
	}
	return fmt.Errorf(
		"%s apunta a la base %q, que no se llama como una base de prueba.\n"+
			"Los tests TRUNCAN todas las tablas. Si esa es tu base de desarrollo, ibas a perderla.\n"+
			"Levantá una aparte y volvé a intentar:\n\n"+
			"  docker run -d --rm --name arena-pg-test -e POSTGRES_PASSWORD=test \\\n"+
			"    -e POSTGRES_DB=arena_test -p 55433:5432 postgres:16-alpine\n\n"+
			"  %s=\"postgres://postgres:test@localhost:55433/arena_test?sslmode=disable\"",
		EnvVar, name, EnvVar)
}

// Pool devuelve un pool con el esquema aplicado y las tablas vacías.
//
// Cada test arranca de cero. Se limpia al principio y no al final: si un test
// falla y deja datos, lo que quedó en la base es evidencia útil para mirarla a
// mano.
func Pool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv(EnvVar)
	if url == "" {
		t.Skipf("sin %s: se saltea el test de base (ver internal/testdb)", EnvVar)
	}
	if err := refuseNonTestDatabase(url); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.Open(ctx, url, db.Options{MaxConns: 6})
	if err != nil {
		t.Fatalf("abriendo la base de prueba: %v", err)
	}
	t.Cleanup(pool.Close)

	// El lock se toma ANTES de migrar, no después. El esquema hace `drop trigger
	// if exists` y `create or replace view`, que piden locks exclusivos sobre las
	// mismas tablas que otro paquete puede estar truncando: migrar sin el turno
	// tomado es el deadlock que esto evita.
	//
	// El lock vive en ESTA conexión, así que se reserva una y se suelta al final
	// del test. Sin timeout en el contexto: esperar el turno puede tardar tanto
	// como tarde el test que lo tiene tomado.
	lockConn, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("tomando la conexión del lock: %v", err)
	}
	if _, err := lockConn.Exec(context.Background(),
		`select pg_advisory_lock($1)`, int64(exclusiveLock)); err != nil {
		lockConn.Release()
		t.Fatalf("tomando el lock de la base de prueba: %v", err)
	}
	t.Cleanup(func() {
		if _, err := lockConn.Exec(context.Background(),
			`select pg_advisory_unlock($1)`, int64(exclusiveLock)); err != nil {
			t.Errorf("soltando el lock de la base de prueba: %v", err)
		}
		lockConn.Release()
	})

	if err := db.Migrate(ctx, pool, quietLogger()); err != nil {
		t.Fatalf("aplicando el esquema: %v", err)
	}
	Truncate(t, pool)
	return pool
}

// Truncate vacía todas las tablas.
//
// TRUNCATE y no DELETE: los triggers de append-only de `coin_transactions` y
// `point_grants` son a nivel de fila y bloquean DELETE, así que un DELETE acá
// fallaría. TRUNCATE no dispara triggers de fila, que es justo lo que hace falta
// para poder limpiar sin poder editar la historia.
func Truncate(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		truncate table
			refresh_tokens,
			point_grants,
			coin_transactions,
			race_settlements,
			race_results,
			bets,
			race_participants,
			horses,
			races,
			invite_codes,
			users
		cascade`)
	if err != nil {
		t.Fatalf("limpiando la base de prueba: %v", err)
	}
}
