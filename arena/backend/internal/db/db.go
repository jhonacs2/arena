// Package db abre el pool de Postgres y aplica el esquema.
//
// Una sola dependencia externa —`github.com/jackc/pgx/v5`— y se usa en modo
// nativo, no a través de `database/sql`: el paquete de carreras necesita pasar
// una `pgx.Tx` en curso al ledger para que el descuento de monedas y la apuesta
// caigan en la MISMA transacción, y por `database/sql` una `pgx.Tx` no se puede
// prestar.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Options son los parámetros del pool que se pueden mover por entorno.
type Options struct {
	// MaxConns es el tope de conexiones abiertas. Supabase cobra conexiones, y
	// el backend es un solo proceso: 10 alcanzan para una clase entera.
	MaxConns int32

	// SimpleProtocol desactiva las sentencias preparadas.
	//
	// Hace falta con el **pooler de transacciones** de Supabase (puerto 6543):
	// ahí cada sentencia puede caer en una sesión distinta, y una sentencia
	// preparada en la sesión A no existe en la B. Con el pooler de sesión o una
	// conexión directa (5432) se deja en false, que es más rápido.
	SimpleProtocol bool
}

// Open arma el pool y verifica que la base contesta.
//
// La cadena de conexión viene entera de DATABASE_URL: cualquier parámetro que
// Postgres entienda —`sslmode`, `search_path`, `application_name`— se pone ahí
// y no hace falta una variable de entorno nueva por cada uno.
func Open(ctx context.Context, url string, opts Options) (*pgxpool.Pool, error) {
	if url == "" {
		return nil, fmt.Errorf("falta DATABASE_URL")
	}

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("DATABASE_URL no se puede interpretar: %w", err)
	}

	if opts.MaxConns > 0 {
		config.MaxConns = opts.MaxConns
	}
	// Una conexión mínima siempre viva: la primera apuesta de la clase no debería
	// pagar el saludo TLS con Supabase.
	config.MinConns = 1

	// Las conexiones se reciclan. Un pooler intermedio puede cortar una conexión
	// vieja sin avisar, y una conexión muerta en el pool aparece como un error
	// aleatorio en la primera petición que la agarra.
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = time.Minute

	if opts.SimpleProtocol {
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("abriendo el pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("la base no contesta: %w", err)
	}
	return pool, nil
}
