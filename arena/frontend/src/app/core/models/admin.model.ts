export interface InviteCode {
  readonly code: string;
  readonly coinsGranted: number;
  readonly note: string | null;
  readonly createdAt: string;
  readonly redeemedAt: string | null;
  readonly redeemedBy: string | null;
}

/** La vista `user_scores`. Es el panel de nota. */
export interface ScoreRow {
  readonly userId: string;
  readonly username: string;
  readonly firstName: string;
  readonly lastName: string;
  readonly balance: number;
  readonly points: number;
  readonly betsPlaced: number;
  readonly betsWon: number;
}

export interface CreateCodesRequest {
  readonly count: number;
  readonly coinsGranted: number;
  readonly note?: string;
}

export interface GiftRequest {
  /** Acepta negativo: es un ajuste, y queda en el ledger con quién lo hizo. */
  readonly coins: number;
  readonly note?: string;
}

export interface GiftResponse {
  readonly balance: number;
  readonly points: number;
}

export interface HorseDraft {
  readonly number: number;
  readonly name: string;
  /** Cuota **nominal** ×100. Informativa: el pago es pari-mutuel, no sale de acá. */
  readonly nominalOdds: number;
}

export interface CreateRaceRequest {
  readonly name: string;
  readonly scheduledAt: string;
  readonly horses: readonly HorseDraft[];
}
