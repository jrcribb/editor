# Revisión de diseño de Editor

> Un espacio de trabajo local para convertir borradores en documentos que puedas
> leer, editar y conservar.

| Revisión | Estado |
| --- | --- |
| Diseño de la interfaz | Listo para la primera revisión |
| Modos de trabajo | Vista previa, dividida, fuente y visual |
| Ubicación de los datos | Archivos locales |

## Una experiencia sencilla

Editor busca que el documento ocupe el centro de la atención. La lista de archivos
y los controles permiten encontrar la siguiente acción con facilidad.

### Tres funciones para el trabajo diario

- **Abre tus archivos** desde el disco y trabaja con ellos en tu equipo.
- **Elige cómo trabajar** con vista previa, código fuente o vista dividida.
- **Conserva tus cambios** guardándolos en el archivo original cuando el navegador lo permita.

> La interfaz debe facilitar la lectura y la edición del contenido.

## Un detalle de implementación

Editor funciona sin conexión desde un único archivo HTML.

```js
const documento = {
  nombre: "Mis notas.md",
  formato: "Markdown",
  ubicacion: "local"
};
```

## Antes de compartir

- [x] Comprobar el funcionamiento sin servidor
- [x] Revisar los temas claro y oscuro
- [x] Probar una pantalla estrecha
- [ ] Añadir las notas de la revisión final

---

_Documento de ejemplo para comprobar la presentación de Editor._
