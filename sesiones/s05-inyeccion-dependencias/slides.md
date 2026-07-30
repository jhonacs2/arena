---
marp: true
theme: neobrutal
paginate: true
header: 'S5 · Inyección de dependencias'
---

<!-- _class: portada -->

# S5

## Inyección de dependencias

<!--
Módulo Angular · Talento DH 8va.
El guión completo está en guion.md y es un teleprompter. Tecla P para las notas.
-->

---

<!-- _class: bloque -->

# 0:00

## Pregunta de apertura

---

## Dos mostradores, la misma comanda

No son padre e hijo. **Son hermanos.**

# ¿Cómo se hablan?

<!--
90 segundos. Van a decir "los subo al padre", "una variable global", "no sé".
TODAS SIRVEN.

"El que dijo subirlos al padre tiene razón, y es lo que se hace cuando no se
conoce lo de hoy. El problema es que el padre termina guardando algo que no le
importa, solo para que dos hijos se enteren. Y con tres niveles de por medio,
eso se convierte en pasar un dato por cinco componentes que no lo usan."
-->

---

<!-- _class: bloque -->

# 0:05

## Wayground

## de S4

<!--
sesiones/s04-directivas-pipes/wayground.csv. Máximo 30 segundos por pregunta.
-->

---

<!-- _class: bloque -->

# 0:12

## El concepto

## Editor cerrado

---

## Nadie construye lo que necesita: lo pide

**Inyector** · quien guarda las instancias y las entrega

<!--
TRES MINUTOS.

"Si un componente escribiera new OrderService(), cada componente tendría el
suyo, con su propia comanda. No habría forma de que vieran lo mismo."

"Inyección de dependencias quiere decir que un objeto NO SE CONSTRUYE lo que
necesita: se lo pide a alguien."

Y la ventaja, que no es escribir menos:
"QUIEN PIDE NO DECIDE CUÁNTOS HAY. Esa decisión se toma en otro lado, en una
línea, y se puede cambiar sin tocar a nadie."
-->

---

## Dónde se declara decide cuántos hay

![w:900](diagramas/donde-vive-un-servicio.svg)

<!--
CINCO MINUTOS sobre el diagrama. Es la sesión.

Panel izquierdo: "providedIn: 'root' es UNA SOLA INSTANCIA para toda la
aplicación. Los dos mostradores reciben EXACTAMENTE EL MISMO OBJETO. No se
copian los datos: es el mismo."

Panel derecho: "providers en el componente es UNO POR CADA COPIA de ese
componente. Dos mostradores, dos cuadernos, y ninguno ve el del otro."

Y abajo, el token: "un servicio se pide por su clase. Una cadena no: no se
puede escribir inject(string). Un InjectionToken es una llave con nombre y con
tipo, hecha para eso."

Si preguntan por qué no una constante exportada: "una constante se importa, y
quien la importa queda atado. Un token se reemplaza en un solo lugar. En la
clase 7, cambiar entre el servidor de verdad y el mock va a ser esa línea."
-->

---

<!-- _class: ojo -->

# ¿Cuántos de estos tiene que haber?

Una comanda: una.

Un cuaderno de mostrador: uno por mostrador.

<!--
LA PREGUNTA QUE DECIDE. Déjala en pantalla unos segundos.

"La respuesta a esa pregunta es la línea que hay que escribir. Y no hay ningún
error que te avise si la contestas mal."
-->

---

<!-- _class: bloque -->

# 0:20

## Live coding

## Ustedes miran

<!--
Proyecto: lab/demo, ruta /s05. Secuencia completa en mision-profe.md.
-->

---

<!-- _class: codigo -->

## El servicio compartido · 0:20

```ts
@Injectable({ providedIn: 'root' })
export class OrderService {
  private readonly _orders = signal<readonly Order[]>([]);
  readonly orders = this._orders.asReadonly();

  add(customer: string, coffee: string): void { … }
}
```

<!--
MUEVE LA COMANDA TAL CUAL. No es una reescritura: es cortar y pegar.

"El estado se escribe IGUAL que la clase pasada: un signal privado y un
computed. UN STORE NO ES UN PATRÓN NUEVO: es lo de S3 con providedIn: 'root'
encima."

Detente en las dos líneas nuevas:
"El signal se escribe adentro; afuera se expone de solo lectura. Nadie puede
llamar a set desde un componente. Para cambiar la comanda hay que pasar por
add(), y LOS MÉTODOS SON EL CONTRATO."

Y la pantalla queda igual — dilo: "no ganamos nada todavía."
-->

---

<!-- _class: ojo -->

# Aparece en el otro mostrador

Sin un solo `input()`. Sin un solo `output()`.

<!--
0:24 — EL MOMENTO DE LA CLASE. Toma un pedido en el A y espera un segundo
antes de hablar.

"No hay ni un input ni un output entre los dos. NO SE CONOCEN. Comparten la
comanda porque los dos le pidieron lo mismo al mismo inyector, y el inyector
les dio EL MISMO OBJETO. No se copiaron los datos: es el mismo."

Y la conexión con S2:
"Esto es lo que en la clase 2 no se podía hacer. input() y output() conectan
padre e hijo. Un servicio conecta a cualquiera con cualquiera."
-->

---

<!-- _class: codigo -->

## El servicio por componente · 0:27

```ts
@Injectable()                       // ← sin providedIn
export class NotepadService { … }
```

```ts
@Component({ providers: [NotepadService] })
export class CounterComponent { … }
```

<!--
Anota algo en el A: no aparece en el B. "Esa única línea es toda la diferencia."

ROTURA DELIBERADA 1 — quita el providers sin poner nada:
R3InjectorError[NotepadService -> NotepadService]: NullInjectorError:
No provider for NotepadService!

"Léelo de derecha a izquierda: no hay proveedor. El servicio existe y compila;
falta que ALGUIEN HAYA DICHO DÓNDE VIVE."

Y ahora ponle providedIn: 'root' y anota otra vez: los dos comparten cuaderno.
DEJA QUE LO MIREN DOS SEGUNDOS.
"Y esto NO ES UN ERROR. No hay mensaje, la aplicación funciona. Es una decisión
de diseño equivocada, y la única forma de darse cuenta es haberse hecho la
pregunta."
-->

---

<!-- _class: codigo -->

## El token · 0:31

```ts
export const SHOP_NAME = new InjectionToken<string>('Nombre del café', {
  providedIn: 'root',
  factory: () => 'Café Compilado',
});
```

```ts
providers: [{ provide: SHOP_NAME, useValue: 'Otro café' }]
```

<!--
"factory es el valor por defecto: con eso funciona sin configurar nada."

"Y esta es la razón de que sea un token y no una constante exportada: una
constante SE IMPORTA, y quien la importa queda atado. Esto se reemplaza en un
solo lugar, sin tocar ni un componente."
-->

---

<!-- _class: codigo -->

## `inject()` donde no va · 0:33

```ts
protected take(): void {
  const service = inject(OrderService);   // ← aquí no
}
```

## `NG0203: inject() must be called from an injection context…`

<!--
ROTURA DELIBERADA 2. COMPILA. Toca el botón.

"El mensaje es de los buenos: te dice EXACTAMENTE DÓNDE SÍ SE PUEDE. En el
constructor, en una factory, en un campo de la clase —que es donde lo pusimos
al principio— o adentro de runInInjectionContext."

"El motivo es simple: cuando corre un método, Angular ya terminó de construir
el componente y no tiene forma de saber a qué inyector preguntarle.
INJECT() VA ARRIBA, SIEMPRE."
-->

---

<!-- _class: bloque -->

# 0:35

## Misión 1

## Una comanda, dos mostradores

<!--
15 minutos, lab/starter. Enunciado en mision-estudiante-1.md.

DILO ANTES DE LARGAR:
"Empieza por mover la comanda sin cambiar nada más: que la pantalla quede
exactamente igual. El segundo mostrador al final — es la comprobación, no el
ejercicio."

Reloj de pistas: 0:43 y 0:47.
-->

---

## Misión 1 — los cuatro

1. Dos `<app-counter>`, y la ruta `/s05`
2. `OrderService` con **`providedIn: 'root'`**
3. `NotepadService` en **`providers` del mostrador**
4. `SHOP_NAME` como `InjectionToken`

**Las dos comprobaciones:** un pedido se ve en los dos · una anotación, en uno solo

<!--
Déjala en pantalla los quince minutos.

Probar con UN solo mostrador no prueba nada: con una copia, los dos servicios
se comportan igual.
-->

---

<!-- _class: bloque -->

# 0:50

## Comparten pantalla

<!--
PREGUNTAS, NO CORRIGES:
1. "¿Cuántas instancias de este servicio hay ahora mismo en pantalla?"
2. "¿Cómo lo sabes?"
3. "Si mañana hay un tercer mostrador, ¿qué tocas?"
4. "¿Puede alguien de afuera vaciar tu comanda sin pasar por un método?"

Lo más probable: alguien puso providedIn: 'root' en el cuaderno y no lo notó
porque solo probó con un mostrador.
"Probar con una copia esconde exactamente el error que esta clase enseña."
-->

---

<!-- _class: bloque -->

# 1:00

## Descanso

## 10 minutos

---

<!-- _class: bloque -->

# 1:10

## Predice y ejecuta

<!--
Las tres respuestas quedaron como tests permanentes en el spec de s05.
mostrar → 60 segundos → ejecutar → explicar.
-->

---

<!-- _class: codigo -->

## 1 · Declarado en los dos lados

```ts
@Injectable({ providedIn: 'root' })
export class OrderService { … }
```

```ts
@Component({ providers: [OrderService] })
export class CounterComponent { … }
```

<!--
NI "aparece en el B" NI "da error".
HAY DOS INSTANCIAS, Y NO HAY NINGÚN ERROR.

Tomando un pedido en el A: A dice 1, B dice 0, el tablero dice 0. Son TRES
comandas distintas.

"Angular busca primero en el inyector más cercano y sube si no lo encuentra.
providedIn: 'root' NO ES UNA ORDEN: es un valor por defecto. Cualquiera que
declare el servicio más abajo lo tapa."

Y por eso es el peligroso: el síntoma llega semanas después, descrito como
"se me perdieron los datos".
-->

---

<!-- _class: codigo -->

## 2 · `inject()` en un método

```ts
protected take(): void {
  const orders = inject(OrderService);
  orders.add(…);
}
```

<!--
COMPILA, y falla al tocar el botón. NG0203.

Muestra la terminal primero, para que se vea que compiló sin una queja.

"TypeScript no tiene forma de saber desde dónde se va a llamar una función. Es
un error de tiempo de ejecución: pasa el build, pasa la revisión, y se rompe
con un usuario adelante."
-->

---

<!-- _class: codigo -->

## 3 · Un servicio que pide otro servicio

```ts
@Injectable({ providedIn: 'root' })
export class BetStore {
  private readonly races = inject(RaceStore);
}
```

<!--
SÍ, EXACTAMENTE ASÍ.

"EL INYECTOR NO DISTINGUE QUIÉN PREGUNTA. Un componente, una directiva, un pipe
y otro servicio piden igual."

La yapa, si sobra un minuto: "¿y si los dos se inyectaran entre sí?" — es una
dependencia circular, Angular la detecta y falla al construir el primero.
"Casi nunca es un problema de Angular: es la señal de que los dos deberían ser
uno solo, o de que falta un tercero."

CIERRE: "¿cuál les habría costado más encontrar?" → el primero, que no falla.
Misma familia que el output sin escuchar de S2 y el push de S3.
-->

---

<!-- _class: bloque -->

# 1:25

## Misión 2

## El estado sale de la pantalla

<!--
20 minutos, en parejas.

TRES COSAS ANTES DE LARGAR:
1. "El estado del listado se muda TAL CUAL a un RaceStore. No es una
   reescritura: es un cambio de dueño. Los computed de S3 se copian y pegan."
2. "Lo que SE QUEDA es lo de presentación: las etiquetas, los tonos y el
   formato de la hora. Otra vista podría querer otras."
3. "BetStore va a inyectar a RaceStore."
-->

---

## Misión 2 — qué se muda y qué se queda

| Al store | Se queda en la pantalla |
|---|---|
| filtro · búsqueda | `STATUS_LABELS` |
| cuál está abierta | `STATUS_TONES` |
| contadores · visibles | el formato de la hora |

**La pregunta:** ¿otra pantalla necesitaría exactamente esto?

<!--
Déjala en pantalla los veinte minutos.

core/ NO IMPORTA DE features/. Es la regla de dependencias y hoy es media
consigna.
-->

---

<!-- _class: bloque -->

# 1:45

## Code review

<!--
Rúbrica del curso, más las dos preguntas de hoy:

"¿Cuántas instancias de este servicio hay? ¿Y cómo lo sabrías SIN ejecutarlo?"

"¿Alguien de afuera del store puede escribir su estado? Si la respuesta es sí,
el store no es el dueño de nada."

Los cinco errores de todos los años están en correccion.md.
-->

---

<!-- _class: ojo -->

# Ahora se puede testear sin componentes

```ts
const store = TestBed.inject(RaceStore);
store.setFilter('live');
```

<!--
EL CIERRE DE LA SESIÓN.

"Cuando el estado vivía dentro de race-list, para probar el filtro había que
crear el componente, dibujarlo y buscar botones en el DOM. Ahora son dos
líneas."

Y el repaso del recorrido:
"Miren el componente. En la clase 1 tenía todo. Hoy tiene las etiquetas, los
tonos y el formato de la hora — LO QUE DE VERDAD ES DE ESTA PANTALLA. Todo lo
demás encontró su lugar, y cada mudanza fue por el mismo motivo: alguien más lo
iba a necesitar."
-->

---

<!-- _class: bloque -->

# 1:55

## Exit ticket

## y tarea

<!--
Tarea: LÉELA EN VOZ ALTA. El punto 1 —testear el store sin componentes— es la
demostración de lo que se ganó.
-->

---

<!-- _class: portada -->

# Hasta la próxima

## S6 · Reactividad

<!--
El anzuelo:
"La clase que viene: reactividad y observables. Los datos van a dejar de estar
ahí y van a empezar a LLEGAR — y con eso aparece el tiempo, que es lo único que
todavía no tuvimos que manejar."
-->
