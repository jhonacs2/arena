package accounts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrTokenReused es un refresh que ya se había canjeado.
//
// Se distingue de ErrNotFound porque significa otra cosa: un token de un solo
// uso que aparece dos veces es la señal de que alguien tiene una copia. La
// respuesta al alumno es la misma —UNAUTHENTICATED, volvé a entrar— pero además
// se le revocan TODAS las sesiones, porque no se sabe cuál de las dos copias es
// la legítima.
var ErrTokenReused = errors.New("refresh token reusado")

// SaveRefreshToken guarda el hash del token.
//
// Recibe el hash, no el token: el store no debería poder ver un token válido, y
// así un volcado de la base no alcanza para hacerse pasar por nadie.
func (s *Store) SaveRefreshToken(ctx context.Context, tokenHash, userID string, issuedAt, expiresAt time.Time) error {
	_, err := s.Pool.Exec(ctx, `
		insert into refresh_tokens (token_hash, user_id, issued_at, expires_at)
		values ($1, $2, $3, $4)`,
		tokenHash, userID, issuedAt, expiresAt)
	if err != nil {
		return fmt.Errorf("guardando el refresh de %s: %w", userID, err)
	}
	return nil
}

// ConsumeRefreshToken canjea el token y devuelve de quién era.
//
// **De un solo uso**: el `used_at is null` del UPDATE es la condición, así que
// dos peticiones simultáneas con el mismo token no pueden ganar las dos. La fila
// no se borra, se marca — hace falta para poder detectar el reuso.
//
// Un token reusado devuelve ErrTokenReused y **revoca toda la familia** en la
// misma transacción. Es la contramedida estándar: si el token viajó a manos de
// alguien más, la sesión legítima también se corta y el dueño se vuelve a
// autenticar, en vez de quedar compartiendo la cuenta sin saberlo.
func (s *Store) ConsumeRefreshToken(ctx context.Context, tokenHash string, now time.Time) (string, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("abriendo la transacción del refresh: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID string
	var usedAt *time.Time
	var expiresAt time.Time

	err = tx.QueryRow(ctx, `
		select user_id::text, used_at, expires_at
		from refresh_tokens
		where token_hash = $1
		for update`, tokenHash,
	).Scan(&userID, &usedAt, &expiresAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("leyendo el refresh: %w", err)
	}

	if usedAt != nil {
		if _, err := tx.Exec(ctx, `delete from refresh_tokens where user_id = $1`, userID); err != nil {
			return "", fmt.Errorf("revocando las sesiones de %s: %w", userID, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return "", fmt.Errorf("confirmando la revocación de %s: %w", userID, err)
		}
		return userID, ErrTokenReused
	}

	if !expiresAt.After(now) {
		if _, err := tx.Exec(ctx, `delete from refresh_tokens where token_hash = $1`, tokenHash); err != nil {
			return "", fmt.Errorf("borrando el refresh vencido: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return "", fmt.Errorf("confirmando el borrado del refresh vencido: %w", err)
		}
		return "", ErrNotFound
	}

	if _, err := tx.Exec(ctx,
		`update refresh_tokens set used_at = $2 where token_hash = $1`, tokenHash, now); err != nil {
		return "", fmt.Errorf("canjeando el refresh: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("confirmando el canje del refresh: %w", err)
	}
	return userID, nil
}

// RevokeRefreshTokens corta todas las sesiones del usuario. Es el logout.
func (s *Store) RevokeRefreshTokens(ctx context.Context, userID string) error {
	if _, err := s.Pool.Exec(ctx, `delete from refresh_tokens where user_id = $1`, userID); err != nil {
		return fmt.Errorf("revocando las sesiones de %s: %w", userID, err)
	}
	return nil
}

// PurgeRefreshTokens borra los vencidos y los ya canjeados que quedaron viejos.
//
// Se corre en el arranque. Los canjeados se guardan un rato después de vencer
// para que la detección de reuso siga funcionando sobre un token recién robado;
// pasado el vencimiento ya no sirve de nada y solo ocupa lugar.
func (s *Store) PurgeRefreshTokens(ctx context.Context, before time.Time) (int64, error) {
	tag, err := s.Pool.Exec(ctx, `delete from refresh_tokens where expires_at < $1`, before)
	if err != nil {
		return 0, fmt.Errorf("limpiando los refresh vencidos: %w", err)
	}
	return tag.RowsAffected(), nil
}
