# SNN · «TEMA DE LA SESIÓN» — guión

> Documento maestro de la sesión. Todo lo demás (slides, misiones, quiz) sale de acá.
> **Antes de dar la clase:** cronometrar contra estos bloques. Si algo se pasa 3 minutos, se recorta acá y no en vivo.

| | |
|---|---|
| **Concepto único** | *Una sola frase. Si no entra en una, la sesión tiene dos temas y hay que partirla.* |
| **Al final saben** | *Tres verbos concretos: "crear un X", "explicar por qué Y", "detectar Z".* |
| **Requisito previo** | S(NN−1) · *qué tienen que traer sabido* |
| **Archivos que se tocan** | `lab/…` · `project/frontend/starter/…` |

---

## 0:00 — Pregunta de apertura · 5 min

> **«…»**

Responden **en el chat**. Sin juicio, sin corregir, sin "casi". Se leen dos o tres en voz alta y se sigue.

No es una evaluación: es para que hablen antes de que empiece la clase. El que escribió algo a los 2 minutos participa el resto de la sesión.

- [ ] Pregunta abierta, sin respuesta correcta
- [ ] Se responde en una línea
- [ ] Se conecta con el concepto de hoy, sin adelantarlo

---

## 0:05 — Wayground de S(NN−1) · 7 min

Correr `sesiones/s(NN-1)-*/wayground.csv`. **De la sesión anterior, no de la de hoy.**

Después de cada pregunta fallada por más de un tercio del curso: 30 segundos de por qué. No más — si necesita más, va a la tarea.

| Pregunta | Si falla mucho, decir |
|---|---|
| *…* | *…* |

---

## 0:12 — Concepto en diagrama · 8 min

**El editor está cerrado.** No hay VS Code en pantalla. Solo el diagrama.

Diapositivas: `slides.md` · Diagrama: `diagramas/….svg`

El objetivo es que puedan dibujar el concepto en una servilleta antes de escribirlo. Si arranca el código acá, copian sintaxis sin modelo mental.

**Analogía de hoy:** *…*

---

## 0:20 — Live coding narrado · 15 min

**Ellos no copian. Miran.** Decirlo explícito al empezar: *"cierren el editor, esto lo vamos a hacer juntos en 15 minutos"*.

Proyecto: `lab/solution` → ruta `/sNN`

| Paso | Qué escribo | Qué narro mientras escribo |
|---|---|---|
| 1 | *…* | *…* |
| 2 | *…* | *…* |
| 3 | *…* | *…* |

**Romperlo a propósito una vez** y arreglarlo en vivo. Que vean el mensaje de error real antes de verlo solos.

---

## 0:35 — Misión 1 · 15 min

Enunciado: `mision-estudiante-1.md` · Trabajan en `lab/starter`, ruta `/sNN` · **Individual**

**Estás en silencio.** Disponible si preguntan, pero no circulás ofreciendo ayuda. Los 15 minutos de pelearse solo con el error son la clase.

Si a los 8 minutos más de la mitad está trabada en lo mismo → una pista para todos, en voz alta, sin resolver.

---

## 0:50 — Dos alumnos comparten pantalla · 10 min

**Preguntás, no corregís.** Aunque esté mal. Aunque duela.

Preguntas que sirven, en este orden:

1. ¿Qué esperabas que pasara acá?
2. ¿Qué pasó?
3. ¿Cómo lo averiguaste?
4. Si tuvieras que explicarle esta línea a alguien que no estuvo hoy, ¿qué le decís?

Elegir **una solución que funciona y una que no**. La que no funciona enseña más, y hay que pedirle permiso a la persona antes.

---

## 1:00 — Descanso · 10 min

Cámara y micrófono apagados. Volver puntual: los 15 minutos de después son los más densos.

---

## 1:10 — Predice y ejecuta · 15 min

Carpeta: `predice-y-ejecuta/` · Respuestas: `predice-y-ejecuta/respuestas.md`

Para cada snippet, en este orden y sin saltearse pasos:

1. Se muestra el código. **No se ejecuta.**
2. *"¿Qué va a pasar? Escribilo en el chat."* — 60 segundos.
3. Se ejecuta.
4. Se explica la diferencia entre lo que dijeron y lo que pasó.

El paso 2 es todo el ejercicio. Ejecutar primero lo convierte en una demo.

| # | Qué está roto | Qué predicen casi todos | Qué pasa de verdad |
|---|---|---|---|
| 1 | *…* | *…* | *…* |
| 2 | *…* | *…* | *…* |
| 3 | *…* | *…* | *…* |

---

## 1:25 — Misión 2, en parejas · 20 min

Enunciado: `mision-estudiante-2.md` · Trabajan en `project/frontend/starter` · **En parejas**

Acá el concepto aterriza en el proyecto ancla. Es el único bloque donde tocan el hipódromo.

Circulás entre las parejas. Escuchás más de lo que hablás. Si una pareja terminó, se le da la extensión que está al final del enunciado — no se la deja mirando.

- [ ] El enunciado dice **qué**, no **cómo**
- [ ] Tiene criterios de listo verificables por ellos mismos
- [ ] Entra en 20 minutos con margen para trabarse una vez

---

## 1:45 — Code review en vivo · 10 min

Una solución de la Misión 2, con permiso de la pareja. Se revisa con la rúbrica de `docs/curriculum.md`, en voz alta y en este orden:

1. ¿`standalone: true` y `OnPush`?
2. ¿Actualiza el estado sin mutar?
3. ¿`any`, `console.log`, imports sin usar?
4. ¿Maneja cargando, vacío y error?
5. ¿El nombre dice lo que la cosa hace?
6. ¿Respeta la regla de dependencias?

Los puntos 1 a 3 los verifica `scripts/verify.mjs`. **Del 4 al 6 los verifica una persona — por eso existe este bloque.**

Empezar por algo que está bien hecho. Siempre hay algo.

---

## 1:55 — Exit ticket y tarea · 5 min

Formulario: `exit-ticket.md` — tres preguntas, se responde en 3 minutos.

1. *Una de recordar.*
2. *Una de aplicar.*
3. **¿Qué quedó confuso?** ← esta es la que importa: es lo primero de la sesión siguiente.

Tarea: `tarea.md`. Se lee el enunciado en voz alta antes de cortar. Si se manda solo por chat, no se hace.

---

## Después de la clase

- [ ] Leer los exit tickets. Lo confuso arranca la próxima sesión.
- [ ] Escribir `wayground.csv` de **esta** sesión — se corre la próxima.
- [ ] Anotar acá abajo qué bloque se pasó de tiempo y por qué.

### Notas de la corrida real

*Se completa después de dar la clase. Es lo que hace que S(NN+1) salga mejor.*
