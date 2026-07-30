package accounts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/talentodh/arena/internal/contract"
	"github.com/talentodh/arena/internal/invite"
)

// Topes de la generación por lote. 200 es más de lo que entra en cualquier
// cohorte y el tope existe para que un `count` mal escrito no llene la tabla.
const (
	MaxCodesPerBatch = 200
	DefaultCoins     = 1000
)

// CheckCode valida el código SIN canjearlo, para habilitar el resto del
// formulario de registro.
//
// Devuelve códigos de error distintos para «no existe» y «ya lo usaron» a
// propósito: el alumno tiene que poder distinguir «lo escribí mal» de «ya me
// registré» (api.md).
func (s *Store) CheckCode(ctx context.Context, code string) (int64, error) {
	var coins int64
	var redeemedBy *string

	err := s.Pool.QueryRow(ctx,
		`select coins_granted, redeemed_by::text from invite_codes where code = $1`,
		code,
	).Scan(&coins, &redeemedBy)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return 0, contract.Errorf(contract.CodeCodeNotFound)
	case err != nil:
		return 0, fmt.Errorf("consultando el código %s: %w", code, err)
	case redeemedBy != nil:
		return 0, contract.Errorf(contract.CodeCodeAlreadyRedeemed)
	}
	return coins, nil
}

// CreateCodes genera un lote y lo guarda.
//
// El `on conflict do nothing` con reintento es lo que resuelve la colisión: con
// 959 millones de combinaciones repetir uno es rarísimo, pero «rarísimo» en el
// arranque de una clase es un error 500 que nadie va a poder explicar. Insertar
// y mirar si entró cuesta lo mismo que confiar.
func (s *Store) CreateCodes(ctx context.Context, adminID string, count int, coins int64, note string) ([]string, error) {
	if count < 1 || count > MaxCodesPerBatch {
		return nil, contract.FieldErrors(map[string]string{
			"count": fmt.Sprintf("Pedí entre 1 y %d códigos.", MaxCodesPerBatch),
		})
	}
	if coins < 1 {
		return nil, contract.FieldErrors(map[string]string{
			"coinsGranted": "Las monedas del código tienen que ser al menos 1.",
		})
	}

	codes := make([]string, 0, count)
	// Presupuesto de intentos: alcanza para varias colisiones seguidas y corta
	// antes de girar para siempre si la base rechaza por otro motivo.
	for attempts := 0; len(codes) < count && attempts < count*8+32; attempts++ {
		code, err := invite.Generate()
		if err != nil {
			return nil, err
		}

		tag, err := s.Pool.Exec(ctx, `
			insert into invite_codes (code, coins_granted, note, created_by)
			values ($1, $2, $3, $4)
			on conflict (code) do nothing`,
			code, coins, nullable(note), adminID)
		if err != nil {
			return nil, fmt.Errorf("guardando el código %s: %w", code, err)
		}
		if tag.RowsAffected() == 1 {
			codes = append(codes, code)
		}
	}

	if len(codes) < count {
		return nil, fmt.Errorf("solo se pudieron generar %d de %d códigos sin repetir", len(codes), count)
	}
	return codes, nil
}

// Code es un código con su estado de canje, como lo ve el instructor.
type Code struct {
	Code         string     `json:"code"`
	CoinsGranted int64      `json:"coinsGranted"`
	Note         string     `json:"note,omitempty"`
	Redeemed     bool       `json:"redeemed"`
	RedeemedBy   string     `json:"redeemedBy,omitempty"`   // el usuario, para poder leerlo
	RedeemedByID string     `json:"redeemedById,omitempty"` // el id, para poder regalarle
	RedeemedAt   *time.Time `json:"redeemedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// ListCodes devuelve todos los códigos, los sin canjear primero.
//
// Ese orden es el que sirve en el aula: lo que se busca es el próximo código
// libre para dictarle a alguien que llegó tarde.
func (s *Store) ListCodes(ctx context.Context) ([]Code, error) {
	rows, err := s.Pool.Query(ctx, `
		select c.code,
		       c.coins_granted,
		       coalesce(c.note, ''),
		       c.redeemed_by is not null as redeemed,
		       coalesce(u.username, ''),
		       coalesce(c.redeemed_by::text, ''),
		       c.redeemed_at,
		       c.created_at
		from invite_codes c
		left join users u on u.id = c.redeemed_by
		order by (c.redeemed_by is not null), c.created_at desc, c.code`)
	if err != nil {
		return nil, fmt.Errorf("listando los códigos: %w", err)
	}
	defer rows.Close()

	items := []Code{}
	for rows.Next() {
		var c Code
		if err := rows.Scan(&c.Code, &c.CoinsGranted, &c.Note, &c.Redeemed,
			&c.RedeemedBy, &c.RedeemedByID, &c.RedeemedAt, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("leyendo un código: %w", err)
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recorriendo los códigos: %w", err)
	}
	return items, nil
}

// nullable manda NULL en vez de cadena vacía para las columnas nulables.
func nullable(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
