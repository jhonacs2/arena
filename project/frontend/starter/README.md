# Hipódromo — proyecto ancla

Simulador de apuestas de carreras de caballos. El saldo es **virtual**: no se mueve dinero real en ningún momento.

Este es el proyecto que vas a construir a lo largo de las 11 sesiones del módulo.

---

## Arrancar

```bash
npm install
npm start
```

Abrí <http://localhost:4200>. Va a abrir en **`/sistema`**: la muestra del sistema de diseño.

Esa pantalla no es parte del producto — es tu referencia. Están todos los colores, las tipografías, los botones, los estados de carga y las sedas de los caballos. Cuando algo se vea raro, mirá ahí primero.

## Los comandos que vas a usar

```bash
npm start        # servidor de desarrollo, se recarga solo
npm run build    # compilar para producción
npm test         # correr los tests
```

---

## Cómo encontrar tu trabajo

Los ejercicios están marcados con un comentario que dice qué sesión son:

```ts
// TODO(S3): filtrar las carreras por estado
```

Buscá `TODO(S3)` en el editor y vas a ver todo lo de esa sesión. **Todo lo que no está marcado ya funciona**: no hace falta que lo toques.

## Antes de entregar

- [ ] `npm run build` pasa sin errores
- [ ] La consola del navegador no tiene errores en rojo
- [ ] La vista muestra algo razonable mientras carga, si está vacía y si falla

---

## Las cuentas de prueba

Contraseña para todas: `Carrera123!`

| Correo | Para qué sirve |
|---|---|
| `ana@hipodromo.test` | La de siempre. Tiene historial y apuestas pendientes. |
| `caro@hipodromo.test` | **Correo sin verificar** — sirve para probar qué pasa cuando alguien intenta apostar sin verificar. |
| `hugo@hipodromo.test` | Saldo 980. Sirve para probar el error de saldo insuficiente. |

## Cómo está organizado

```
src/app/
├── core/       cosas que existen una sola vez: modelos, servicios, datos de prueba
├── shared/     piezas reutilizables: botón, seda, estados de carga
├── layout/     la cabecera y el marco de la app
└── features/   las pantallas
```

Una regla que se evalúa: **`features/` puede usar `core/` y `shared/`, pero `shared/` no usa nada de las otras dos.** Si te encontrás necesitando romperla, mové el código en vez de romperla.
