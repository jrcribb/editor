/**
 * Deriva una clave criptográfica de 256 bits a partir de una semilla (texto) y un salt.
 */
async function derivarClave(semilla, salt) {
  const enc = new TextEncoder();
  const materialClave = await crypto.subtle.importKey(
    "raw",
    enc.encode(semilla),
    { name: "PBKDF2" },
    false,
    ["deriveKey"]
  );

  return crypto.subtle.deriveKey(
    {
      name: "PBKDF2",
      salt: salt,
      iterations: 100000, // Alta seguridad contra fuerza bruta
      hash: "SHA-256"
    },
    materialClave,
    { name: "AES-GCM", length: 256 },
    false,
    ["encrypt", "decrypt"]
  );
}

/**
 * Encripta un texto usando una semilla. Devuelve un string en Base64.
 */
async function encriptar(textoPlano, semilla) {
  const enc = new TextEncoder();
  
  // 1. Generar valores aleatorios únicos para esta encriptación
  const salt = crypto.getRandomValues(new Uint8Array(16));
  const iv = crypto.getRandomValues(new Uint8Array(12)); // Vector de inicialización

  // 2. Obtener la clave desde la semilla
  const clave = await derivarClave(semilla, salt);

  // 3. Encriptar el contenido
  const encriptado = await crypto.subtle.encrypt(
    { name: "AES-GCM", iv: iv },
    clave,
    enc.encode(textoPlano)
  );

  // 4. Combinar salt + iv + datos encriptados en un solo arreglo
  const bufferTotal = new Uint8Array(salt.byteLength + iv.byteLength + encriptado.byteLength);
  bufferTotal.set(salt, 0);
  bufferTotal.set(iv, salt.byteLength);
  bufferTotal.set(new Uint8Array(encriptado), salt.byteLength + iv.byteLength);

  // 5. Convertir a Base64 para facilitar su almacenamiento o transporte
  return btoa(String.fromCharCode(...bufferTotal));
}

/**
 * Función inversa: Desencripta el texto en Base64 usando la misma semilla.
 */
async function desencriptar(textoBase64, semilla) {
  try {
    // 1. Decodificar de Base64 a bytes
    const binario = atob(textoBase64);
    const bufferTotal = new Uint8Array(binario.length);
    for (let i = 0; i < binario.length; i++) {
      bufferTotal[i] = binario.charCodeAt(i);
    }

    // 2. Extraer salt (primeros 16 bytes), iv (siguientes 12) y los datos
    const salt = bufferTotal.slice(0, 16);
    const iv = bufferTotal.slice(16, 28);
    const datosEncriptados = bufferTotal.slice(28);

    // 3. Reconstruir la misma clave usando el salt extraído y la semilla
    const clave = await derivarClave(semilla, salt);

    // 4. Desencriptar
    const desencriptado = await crypto.subtle.decrypt(
      { name: "AES-GCM", iv: iv },
      clave,
      datosEncriptados
    );

    // 5. Convertir bytes de vuelta a texto legible
    const dec = new TextDecoder();
    return dec.decode(desencriptado);
  } catch (error) {
    throw new Error("Fallo al desencriptar: semilla incorrecta o datos corruptos.");
  }
}
