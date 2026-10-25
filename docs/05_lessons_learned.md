# 05 — Lecciones Aprendidas

> Incidentes, decisiones y aprendizaje continuo del proyecto.

---

## 2026-07-30 — Finalización: Observer + ATENEA + Update Checker

### Lección 1: PowerShell `@()` array forcing es obligatorio con git
**Problema:** `git tag --list` devuelve un solo string cuando hay un tag. Indexar con `[0]` da el primer **carácter**, no el tag completo (ej: "v" en vez de "v4.2.0").
**Solución:** Forzar array con `@(git tag --list)` para garantizar que siempre sea un array, incluso con un solo resultado.

### Lección 2: `git describe --tags` no es confiable para detectar último tag
**Problema:** `git describe --tags origin/main` devuelve formato `v4.2.0-15-gdc1e74b` (tag + commit count + hash) cuando HEAD no apunta directo a un tag.
**Solución:** Usar `git fetch --tags` + `git tag --list | Sort-Object` para obtener tags limpios.

### Lección 3: PowerShell no tiene `String.TrimStart()` en chars
**Problema:** `$v.TrimStart('v')` funciona en C# pero no en PowerShell cuando `$v` es un char en vez de string (por el bug de arrays).
**Solución:** Solucionar el bug de raíz (forzar array) y usar `$v.TrimStart('v')` contra strings reales.

### Lección 4: Go embed + SSE es perfecto para dashboards single-binary
**Decisión:** `//go:embed` + Server-Sent Events elimina necesidad de WebSocket, proxy reverso, o servidor de archivos externo.
**Resultado:** `mvgn-scan.exe` sigue siendo single-binary (~6MB), actualizable por reemplazo.

### Lección 5: ATENEA scoring debe ser un blend, no reemplazo
**Decisión:** ATENEA contribuye al 20% del health score final. Si fuera 100% predominaría sobre el estado real del proyecto.
**Fundamento:** El status baseline (completed=95, in_progress=60, etc.) representa el progreso real; ATENEA añade contexto de governance sin opacarlo.

---

## 2026-07-30 — Observation Console MVP

### Lección 1: El blackboard ya existía (no UPGRADE)
**Problema:** Al iniciar sesión, el detector indicó modo UPGRADE porque no encontró `.mvgn/blackboard/state.json` con el glob inicial.
**Realidad:** El blackboard SÍ existía. El glob `**/.mvgn/blackboard/state.json` falló porque `.mvgn/` es un directorio oculto y el patrón no lo atravesó correctamente.
**Solución:** Usar `Get-ChildItem -Recurse` para verificar existencia real. El proyecto estaba en modo READY, no UPGRADE.

### Lección 2: `%!s(MISSING)` en Go
**Problema:** Al renderizar el task ID aparecía `audit-athn1-obs-console-001%!s(MISSING)`.
**Causa:** `display.Field()` ya devuelve el string completo con indentación y valor. Al pasarlo a `fmt.Sprintf("  %s%s", ...)` con un solo argumento, el segundo `%s` quedaba sin argumento.
**Solución:** Usar solo `display.Field()` sin wrappers adicionales.

### Lección 3: ANSI escapes en MetricCard
**Problema:** `strings.Repeat` con count negativo por ANSI codes en el value.
**Causa:** `MetricCard` usaba `len(value)` que cuenta bytes ANSI. La health string con colores ANSI media ~27 chars visibles pero ~35 bytes totales.
**Solución:** Implementar `visibleLen()` que cuenta solo caracteres visibles ignorando secuencias ANSI.

### Lección 4: Windows console y doble click
**Problema:** El .exe se cerraba inmediatamente al hacer doble click.
**Causa:** Sin flag `-watch` o `-observe`, el modo default es scan-once-and-exit. La ventana de consola se abre y cierra en milisegundos.
**Solución:** Crear `tools/observe.bat` con `-observe` + `pause` para doble click directo.

### Lección 5: Go `embed` es perfecto para dashboards single-binary
**Decisión:** Usar `//go:embed` para incrustar HTML+CSS+JS en el binario.
**Fundamento:** Zero dependencias externas, distribución como `.exe` único, sin necesidad de servir archivos desde disco.
**Resultado:** `mvgn-scan.exe` pasó de ~4MB a ~6MB, sigue siendo un solo archivo.

---

## 2026-07-30 — Implementación de Audio Ingestion v5.2

### Lección 1: No saltarse el ciclo MVGN
**Problema:** La implementación del PRD de ATHN1 se hizo directa sin generar PRD → Arquitectura → Tasks previo.
**Consecuencia:** Hubo que generar la documentación de manera inversa.
**Solución:** Siempre seguir el pipeline MVGN: PRD → Arquitectura → Tasks → Implementación → Documentación.

### Lección 2: Faster-Whisper vs OpenAI Whisper
**Decisión:** Faster-Whisper como motor default por ser 4x más rápido y soportar INT8 quantization.
**Riesgo detectado:** OpenAI Whisper requiere `ffmpeg` en el PATH del sistema. El script no verifica esto.
**Mitigación:** Agregar verificación de `ffmpeg` en futura iteración cuando `-Engine whisper`.

### Lección 3: Dependencias Python en Windows
**Incidente:** `pip install faster-whisper` puede tardar >30s porque descarga torch (2GB+).
**Solución:** El script `transcribe.ps1` maneja la instalación con flag `-InstallDeps` y timeout largo.
**Recomendación:** Documentar que la primera instalación requiere descarga grande.

### Lección 4: context.md como filtro semántico
**Decisión:** El PRD debe usar exclusivamente `context.md`, no `transcript.md` crudo.
**Fundamento:** `transcript.md` contiene ruido (muletillas, timestamps, solapamiento). `context.md` extrae solo lo relevante (objetivos, requisitos, riesgos).
**Refuerzo:** Mensaje explícito al final del pipeline advirtiendo esta regla.

---

*Documento generado por MVGN — v5.2*
