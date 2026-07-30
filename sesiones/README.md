# Sesiones

Una carpeta por sesión, todas con los mismos archivos. La plantilla vacía está en `_plantilla/`.

```
sesiones/
├── _plantilla/            copiar de acá para arrancar una sesión nueva
├── s00-typescript/        tema 0, asíncrono, antes de S1
├── s01-primer-componente/
├── …
└── s11-produccion/
```

## Los archivos de cada sesión

| Archivo | Bloque del guión | Quién lo lee | Qué es |
|---|---|---|---|
| `guion.md` | todos | instructor | Los 12 bloques con sus minutos y qué decir en cada uno. **Es el documento maestro.** |
| `slides.md` | 0:12 en adelante | ambos | Marp. El guión de cada diapositiva vive en sus speaker notes. |
| `diagramas/*.excalidraw` | 0:12 | ambos | El concepto dibujado. Editor cerrado. El `.svg` de al lado es generado. |
| `mision-profe.md` | 0:20 | instructor | **La secuencia de tecleo del live coding**, en `demo/`. |
| `mision-estudiante-1.md` | 0:35 | alumno | Ejercicio individual, en `lab/starter`. |
| `mision-estudiante-2.md` | 1:25 | alumno | Ejercicio en parejas, en `project/frontend/starter`. |
| `correccion.md` | 1:45 | ambos | **De vacío a funcionando, paso a paso.** Es lo que hace posible el starter vacío. |
| `conceptos.md` | después de clase | alumno | **El apunte de referencia.** Lo que reemplaza a la memoria en la tarea. |
| `predice-y-ejecuta/` | 1:10 | ambos | Snippets rotos + `respuestas.md`. |
| `wayground.csv` | 0:05 de la **sesión siguiente** | instructor | Preguntas sobre **esta** sesión. |
| `exit-ticket.md` | 1:55 | ambos | Tres preguntas. |
| `tarea.md` | 1:55 | alumno | Consigna asíncrona. |

El nombre de las misiones dice **quién teclea**: `mision-profe` es lo que se demuestra en vivo, `mision-estudiante-N` lo que codean ellos. Los ejercicios de alumno se escriben como ejercicio de libro —enunciado, datos, requisitos, resultado esperado, restricciones—, sin voz de instructor: se leen en casa, sin el contexto del aula.

## El desfase del Wayground

> `sesiones/sNN/wayground.csv` tiene preguntas **sobre la sesión NN** y se corre al empezar la **sesión NN+1**.

Es lo que más fácil se rompe al armar los materiales. Dos consecuencias:

- El bloque 0:05 de **S1** usa `s00-typescript/wayground.csv`.
- El quiz de **S11** no se corre en clase: va a la evaluación asíncrona de cierre.

## Generar las diapositivas

```bash
npx @marp-team/marp-cli --theme-set theme/ sesiones/s01-*/slides.md -o slides.html
npx @marp-team/marp-cli --theme-set theme/ sesiones/s01-*/slides.md --pdf --allow-local-files
```

Los `.html` y `.pdf` generados están en `.gitignore`: se regeneran, no se versionan.
