# 04 — Changelog

```
DOC-REF: CHANGELOG/1.0
SIG: Change Record
DIVISION: MVGN Labs — Research & Creative Systems Division
NODE: MX-03
ΚΛΕΙΣ: CHNG-001
STATE: ACTIVE
```

> Formato: [Keep a Changelog](https://keepachangelog.com) + Semantic Versioning.

---

## [5.3.0] - 2026-08-26 - Separation: dev/impl separation + LAYOUT/2.0

### Added
- **Distribucion por GitHub Releases**: tools/bootstrap.ps1 (instalador una linea irm|iex), tools/package-release.ps1 (zip por manifiesto + validacion 1:1)
- **LAYOUT/2.0**: todo el motor MVGN consolidado en .mvgn/ - publicacion segura por construccion (.gitignore de 1 linea para adoptantes)
- **JOIN/2.0**: protocolo canonico inter-modelo con cosmogonia CNX (designadores AAAA-##, panteon, custodia AEON) + Ciclo de supervision estandar COLLAB
- **templates/**: docs base 00-04 como fuente unica, payload-manifest.json, gitignore.template, schemas/
- **CLI global mvgn**: tools/install-profile.ps1 - bloque idempotente con resolucion ascendente .mvgn/NYX.ps1
- **Audio ingestion GPU**: importado y probado (CUDA int8, batch queue, processed.json); punto unico tools/mvgn-audio.ps1; opcion en INIT y READY (-Audio)
- **Guards legacy**: bootstrap aborta ante prototipos pre-5.x y layout pre-consolidacion; cnx.ps1 etiqueta [LEGACY]
- **Interfaz CNX en bootstrap**: marco/banner estilo cnx.ps1, referencias griegas leves (ΠΥΛΗ, ΤΕΛΟΣ, ΕΝΤΑΞΕΙ/ΣΦΑΛΜΑ, sello AEON), barra de carga real por bytes y ENTER obligatorio al cierre (-NoPause para invocacion programatica)

### Changed
- NYX.ps1 -Mode update delega en bootstrap (semver contra Releases, changelog delta; sin git history)
- mvgn-init.ps1 determinista: copia desde templates/, cero here-strings, ProjectRoot derivado de script
- cnx.ps1: flujo clone-based eliminado (3 fugas cerradas)

### Fixed
- H-6: flujo update de bootstrap (una sola descarga/extraccion sirve changelog+deploy)
- H-13: modo -PayloadZip no muta runtime_version
- INC-1 (supervision): motor revertido por test fuera de sandbox - restaurado con danio colateral incluido

---
## [0.1.0] — FECHA

### Added

- Proyecto inicializado con MVGN.

---

## Rito de sellado (releases)

Al publicar un release, su entrada se cierra con sello AEON:

```
ΣΦΡΑΓΙΣ: ΑΠΟΛΥΤΟΣ
ΦΥΛΑΚΗ: AEON-PROTOCOL-v1 :: SEALED
```

Solo versiones verificadas reciben sello. Lo sellado no se reescribe: se corrige en versión nueva.

