# 3 · Dos huecos iguales

**No ejecutes todavía.** Mirá el código y escribí en el chat qué va a pasar.

---

El hijo tiene el mismo `<ng-content>` dos veces, en dos lugares distintos:

```html
<article class="card">
  <h3>{{ coffee().name }}</h3>

  <div class="card__slot">
    <ng-content />
  </div>

  <p class="card__price num">{{ coffee().price }}</p>

  <div class="card__slot">
    <ng-content />
  </div>
</article>
```

El padre proyecta una sola cosa:

```html
<app-coffee-card [coffee]="item.coffee">
  <p class="note">Vuelve el jueves.</p>
</app-coffee-card>
```

### La pregunta

**¿Cuántas veces aparece «Vuelve el jueves.» en la tarjeta: cero, una o dos? ¿O
directamente no compila?**

Escribí tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s02/coffee-card.component.html`, duplicá el bloque
del `<ng-content />` sin `select`.

**Mostrá la terminal primero** —para que se vea que compiló— y recién después el
navegador. Después de explicar, se borra el duplicado.
