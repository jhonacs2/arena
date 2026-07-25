#!/usr/bin/env bash
# Envoltorio de compatibilidad. La verificación real vive en verify.mjs, que es
# multiplataforma: los alumnos están en Windows, macOS y Linux, y un .sh no
# corre igual en los tres.
#
#   ./scripts/verify.sh              todo
#   ./scripts/verify.sh --fast       sin builds
#   ./scripts/verify.sh contrato     un solo grupo
set -e
exec node "$(dirname "$0")/verify.mjs" "$@"
