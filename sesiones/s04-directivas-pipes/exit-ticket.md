# S4 · Exit ticket

**3 minutos. Anónimo si prefieres.**

---

**1. Recordar** — Para cada caso, marca si es un **pipe**, una **directiva de
atributo** o una **directiva estructural**:

| Qué hay que hacer | Qué se usa |
|---|---|
| Mostrar una cuota siempre con dos decimales | |
| Marcar la fila del caballo favorito | |
| Dibujar una estrella por cada punto de puntaje | |
| Poner un nombre en mayúsculas | |

**2. Aplicar** — El template dice `NG8004: No pipe found with name 'odds'`, y el
archivo `odds.pipe.ts` existe y compila. Escribe **las dos cosas** que hay que
revisar.

**3. ¿Qué quedó confuso?**

---

## Para el instructor

**Respuestas de la 1:** pipe · directiva de atributo · directiva estructural ·
pipe.

**Respuesta de la 2:**

1. Que el pipe esté en el array `imports` del componente.
2. Que el `name` del `@Pipe` coincida **exacto** con lo escrito en el template.

Quien conteste solo la primera se llevó la regla; quien conteste las dos entendió
que se comparan **cadenas**, no clases.

| Lo que escriben | Qué decirles |
|---|---|
| «Registrarlo en `app.config.ts`» | Ahí se registran proveedores, no pipes. Un pipe se declara en el componente que lo usa. |
| «Exportarlo desde `index.ts`» | Ayuda a importarlo, pero el error es otro: el archivo ya se está importando. |
| «Ponerle `standalone: true`» | Si faltara, el error sería `TS-992011` y diría otra cosa. Vale mostrar los dos mensajes juntos. |

| Señal | Qué hacer |
|---|---|
| Confunden pipe con directiva de atributo | Abre S5 con el diagrama, 5 minutos |
| Aparece «no entendí el asterisco» | Es esperable: era de lectura. Está entero en `conceptos.md` §6 |
| Nadie confundido | Subir el techo de la Misión 2 la próxima |
