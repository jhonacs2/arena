# Lab — Módulo Angular

Acá se practica **el concepto de cada sesión, aislado**. Sin el ruido del proyecto del hipódromo: un tema por ruta, en un dominio chico y conocido — una cafetería.

Es donde pasan el live coding, la Misión 1 y el bloque de «predice y ejecuta». La Misión 2, la del proyecto ancla, pasa en el otro repo.

---

## Arrancar

```bash
npm install
npm start
```

Abrí <http://localhost:4200>. La barra de la izquierda tiene las 11 sesiones; las que todavía no se dieron están en gris y no se pueden abrir.

## Encontrar tu trabajo

Cada sesión deja sus ejercicios marcados:

```ts
// TODO(S1): interpolar el nombre del café
```

Buscá `TODO(S1)` en el editor y vas a ver todo lo de esa clase. **Lo que no está marcado ya funciona.**

## Comprobar

```bash
npm test
```

Los tests prueban la **lógica de la clase**, que viene hecha. Los bindings del template —que son tu trabajo— los comprobás vos en el navegador, con la lista de «Listo cuando» del enunciado.

Aprender a mirar la pantalla y decidir si está bien es parte del ejercicio; un test verde no reemplaza eso.

---

## Sesiones

| | Tema |
|---|---|
| **S0** | Tipos, uniones, opcionales y genéricos |
| **S1** | Primer componente standalone y los cuatro bindings |
| S2 | `input`, `output` y `ng-content` |
| S3 | Signals y control flow |
| S4 | Directivas y pipes |
| S5 | Inyección de dependencias |
| S6 | Reactividad y observables |
| S7 | `HttpClient` |
| S8 | Reactive Forms |
| S9 | Routing |
| S10 | WebSockets, OnPush y NgModules como legado |
