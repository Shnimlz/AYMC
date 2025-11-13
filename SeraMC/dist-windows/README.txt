═══════════════════════════════════════════════════════════════
    AYMC SeraMC - Advanced Minecraft Control Panel
                    Version 0.1.0
═══════════════════════════════════════════════════════════════

�� Contenido del Paquete
──────────────────────────────────────────────────────────────
✅ seramc.exe         Aplicación principal (21 MB)
✅ README.txt         Este archivo

🚀 Instalación Rápida
──────────────────────────────────────────────────────────────
1. Copiar seramc.exe a cualquier carpeta
2. Doble clic para ejecutar
3. ¡Listo! No requiere instalación adicional

💻 Requisitos del Sistema
──────────────────────────────────────────────────────────────
Sistema Operativo:
  ✓ Windows 10 (64-bit) - Build 1809 o superior
  ✓ Windows 11 (64-bit)

Hardware Mínimo:
  ✓ 4 GB RAM
  ✓ 200 MB espacio en disco
  ✓ Procesador x64

Software:
  ✓ WebView2 Runtime
    - Incluido en Windows 11
    - Auto-descarga en Windows 10
    - Manual: https://go.microsoft.com/fwlink/?linkid=2124701

🔒 Advertencia de Seguridad (Primera Ejecución)
──────────────────────────────────────────────────────────────
Windows SmartScreen puede mostrar:
  "Windows protected your PC"
  "Unknown publisher"

¿Por qué?
  La aplicación no está firmada digitalmente.

Solución:
  1. Clic en "More info"
  2. Clic en "Run anyway"
  3. La aplicación es segura (código abierto)

📖 Uso Básico
──────────────────────────────────────────────────────────────
1. Ejecutar seramc.exe
2. Configurar conexión SSH a tu VPS
3. Instalar backend y agent automáticamente
4. Gestionar servidores Minecraft desde Windows

🌐 Conexión SSH
──────────────────────────────────────────────────────────────
Requisitos en tu VPS/Servidor:
  ✓ SSH activado (puerto 22 o custom)
  ✓ Usuario con sudo (sin contraseña recomendado)
  ✓ Linux (Arch, Ubuntu, Debian, RHEL, CentOS)

Métodos de autenticación soportados:
  ✓ Contraseña
  ✓ Clave privada (archivo .pem, .key)
  ✓ Passphrase para claves

📁 Ubicación de Datos
──────────────────────────────────────────────────────────────
Windows guarda los datos en:
  %APPDATA%\com.shni.aymc.seramc\

Incluye:
  - Configuración de conexiones SSH
  - Logs de la aplicación
  - Cache de sesiones

🛠️ Problemas Comunes y Soluciones
──────────────────────────────────────────────────────────────

❌ "No se puede conectar al VPS"
   ✓ Verificar IP y puerto SSH
   ✓ Verificar firewall del VPS
   ✓ Comprobar credenciales

❌ "WebView2 Runtime no encontrado"
   ✓ Descargar: https://go.microsoft.com/fwlink/?linkid=2124701
   ✓ Instalar y reiniciar aplicación

❌ "Permission denied (publickey)"
   ✓ Verificar que la clave privada es correcta
   ✓ Verificar permisos de la clave
   ✓ Usar passphrase si la clave está encriptada

❌ "Backend no se instala"
   ✓ Verificar que el usuario tiene sudo
   ✓ Verificar conexión a internet en el VPS
   ✓ Revisar logs en /var/log/aymc/

🔗 Enlaces Útiles
──────────────────────────────────────────────────────────────
Repositorio GitHub:
  https://github.com/Shnimlz/AYMC

Documentación Completa:
  https://github.com/Shnimlz/AYMC/blob/main/docs/

Reportar Problemas:
  https://github.com/Shnimlz/AYMC/issues

📝 Características Principales
──────────────────────────────────────────────────────────────
✅ Conexión SSH segura a VPS remotos
✅ Instalación automática de backend y agent
✅ Gestión de servidores Minecraft
✅ Terminal remoto integrado
✅ Monitoreo en tiempo real
✅ Backups automáticos
✅ Marketplace de plugins
✅ Logs avanzados con análisis

🎮 Servidor Minecraft Soportados
──────────────────────────────────────────────────────────────
✓ Vanilla (Java Edition)
✓ Spigot
✓ Paper
✓ Purpur
✓ Fabric
✓ Forge

Versiones:
  1.8.x - 1.21.x

📄 Licencia
──────────────────────────────────────────────────────────────
MIT License
Copyright (c) 2025 AYMC Team

Código abierto y gratuito.
Ver: https://github.com/Shnimlz/AYMC/blob/main/LICENSE

👥 Soporte
──────────────────────────────────────────────────────────────
¿Necesitas ayuda?
  GitHub Issues: https://github.com/Shnimlz/AYMC/issues
  Email: tu@email.com
  Discord: [Tu servidor de Discord]

🎉 ¡Gracias por usar AYMC SeraMC!
──────────────────────────────────────────────────────────────
Build: 2025-01-13
Versión: 0.1.0
Plataforma: Windows x64

