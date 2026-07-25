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

| Archivo | Bloque del guión | Qué es |
|---|---|---|
| `guion.md` | todos | Los 12 bloques con sus minutos y qué decir en cada uno. **Es el documento maestro.** |
| `slides.md` | 0:12 en adelante | Marp. El guión de cada diapositiva vive en sus speaker notes. |
| `diagramas/*.svg` | 0:12 | El concepto dibujado. Editor cerrado. |
| `mision-1.md` | 0:35 | Enunciado individual, en `lab/starter`. |
| `mision-2.md` | 1:25 | Enunciado en parejas, en `project/frontend/starter`. |
| `predice-y-ejecuta/` | 1:10 | Snippets rotos + `respuestas.md`. |
| `wayground.csv` | 0:05 de la **sesión siguiente** | Preguntas sobre **esta** sesión. |
| `exit-ticket.md` | 1:55 | Tres preguntas. |
| `tarea.md` | 1:55 | Consigna asíncrona. |

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
