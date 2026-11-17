// Ejemplo de uso de cifrado.js
async function probarEncriptacion() {
  const mensajeSecreto = "Datos confidenciales: Cuenta #987654321";
  const miSemilla = "mi-semilla-super-secreta-2026";

  console.log("Original:", mensajeSecreto);

  // Encriptar
  const cifrado = await encriptar(mensajeSecreto, miSemilla);
  console.log("Encriptado:", cifrado);

  // Desencriptar (Inversa)
  const descifrado = await desencriptar(cifrado, miSemilla);
  console.log("Desencriptado:", descifrado);
}

probarEncriptacion();
