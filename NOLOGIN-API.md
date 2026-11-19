# Adaptador experimental NoLogin

Inspeccionado y probado el 14 de noviembre de 2026. No es una API oficial ni una garantía de disponibilidad del proveedor.

El editor público importa Firebase Firestore. Lee `doc(firestore,"documents",pageName)` mediante `getDoc` y guarda mediante `setDoc`. La configuración pública identifica el proyecto `no-login-044`. No se accedió a páginas ajenas ni se enumeró la colección.

## Rutas reproducidas

Base: `https://firestore.googleapis.com/v1/projects/no-login-044/databases/(default)/documents/documents`

- Crear: `POST BASE?documentId=ID`, cuerpo `{fields: ...}` con valores tipados de Firestore.
- Leer: `GET BASE/ID`.
- Actualizar texto: `PATCH BASE/ID?updateMask.fieldPaths=text&updateMask.fieldPaths=textTabs&currentDocument.updateTime=VERSION`.

La versión es `updateTime` devuelta por Firestore, conservada como cadena exacta. El adaptador actualiza solo texto y pestañas, conservando otros campos. El editor original reemplaza el documento; no conviene mezclar ediciones concurrentes desde el editor y desde este adaptador.

## Uso

```js
const id = 'torbellino-' + crypto.randomUUID();
await NoLoginAPI.create(id, JSON.stringify({dispositivos: []}), 1);
const nota = await NoLoginAPI.read(id);
const datos = JSON.parse(nota.content);
datos.dispositivos.push({alias: 'Equipo A', id: 'otra-pagina'});
await NoLoginAPI.update(id, JSON.stringify(datos), nota.version);
```

Si devuelve `FAILED_PRECONDITION`, volver a leer y combinar los datos antes de reintentar. No reemplazar ciegamente con la versión anterior. Crear con un ID ocupado devuelve `ALREADY_EXISTS`.

`nologin-api.js` expone `NoLoginAPI.create/read/update`; no necesita SDK ni credenciales. `nologin-api-demo.html` contiene una copia incorporada del cliente y es un archivo autónomo con una prueba manual. La demo crea una nota ficticia por ejecución.

## Validación real

- Creación, lectura, actualización y lectura posterior: HTTP 200.
- Escritura con versión antigua: HTTP 400, `FAILED_PRECONDITION`.
- Cliente JavaScript ejecutado con Node: creación, lectura, actualización, rechazo de duplicado y versión antigua correctos.
- CORS: preflight PATCH aceptado para `https://example.org` y `null`; GET y PATCH devolvieron cabecera del origen correspondiente.
- No se ejecutó la demo en navegador ni se verificó en celulares. La comprobación CORS por HTTP es evidencia favorable, no una prueba completa del entorno final.
- Resultados conservados en `nologin-research/*results.json`. La primera prueba tuvo un error de formato de fecha de PowerShell; la segunda conserva la cadena exacta y verifica la actualización correctamente.

## Límites

Usar solo IDs propios aleatorios, con prefijo `torbellino-`. El cliente rechaza páginas marcadas con contraseña o bloqueos y no implementa su elusión. No incluye listado global, eliminación, archivos ni cambios de permisos.

El código del editor compara algunos controles de contraseña y bloqueo en el cliente. No se verificó que esos controles estén aplicados por el servidor: no confiar en este mecanismo para almacenar claves de acceso sensibles. Para datos privados se necesita cifrado del lado del cliente con un secreto fuerte, y no un código numérico corto.

La caducidad se registra como `expirationTimestamp` y el adaptador rechaza notas vencidas. No se verificó borrado automático puntual del servidor; el código del sitio también contiene limpieza de notas expiradas. No asumir eliminación garantizada al cumplirse una hora.

Es infraestructura de un tercero sin contrato de integración identificado: rutas, permisos y cuotas pueden cambiar. Para un prototipo funciona en las pruebas realizadas; para continuidad y control conviene un proyecto Firebase propio.
