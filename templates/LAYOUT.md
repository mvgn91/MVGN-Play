# MVGN Layout Contract

```
DOC-REF: LAYOUT/2.0
SIG: Estructura Canónica de Proyecto
DIVISION: MVGN Labs — Research & Creative Systems Division
NODE: MX-03
STATE: ACTIVE
ΚΛΕΙΣ: LAYOUT-002
```

> Contrato único de estructura para TODO proyecto que adopta MVGN.
> Regla fundamental: **MVGN vive íntegro en `.mvgn/`. El proyecto es del proyecto.**

---

## 1. Estructura canónica

```
proyecto/
├── .mvgn/                        # ═══ TODO lo de MVGN ═══
│   ├── NYX.md                    # Entry point universal — "Lee .mvgn/NYX.md"
│   ├── NYX.ps1                   # Entry point imperativo
│   ├── MVGN.md                   # Documentación del framework
│   ├── mvgn-runtime.json         # Manifiesto del runtime (nombre intacto por compatibilidad)
│   ├── LAYOUT.md                 # Este contrato, copiado al adoptar
│   ├── blackboard/
│   │   ├── state.json            # Fuente de verdad del estado
│   │   ├── context.md            # Contexto general del dominio
│   │   ├── JOIN.md               # Protocolo JOIN — entrada de agentes
│   │   └── handoff-latest.json   # Último handoff entre activos
│   ├── templates/                # Fuente canónica (docs base, JOIN, gitignore)
│   ├── scripts/                  # Orquestador + audio pipeline [opcional]
│   ├── tools/                    # init, bootstrap, validadores
│   ├── audio/                                                              [runtime]
│   └── logs/
│       └── execution.log                                                   [runtime]
├── docs/                         # ═══ SOLO documentación DEL PROYECTO ═══
│   ├── 00_idea.md                # (desplegadas desde .mvgn/templates/docs)
│   ├── 01_prd.md
│   ├── 02_architecture.md
│   ├── 03_tasks.md
│   └── 04_changelog.md
├── src/                          # Código fuente del proyecto
├── README.md                     # Del proyecto
└── .gitignore                    # Desplegado desde .mvgn/templates/gitignore.template
```

**Leyenda:** `[runtime]` = regenerable, nunca se publica · `[opcional]` = solo si la capability está activa.

## 2. Reglas del contrato

1. **Consolidación:** todo archivo de la herramienta MVGN reside bajo `.mvgn/`.
   Nada de MVGN aparece en la raíz del proyecto. Sin excepciones.
2. **Separación de autoría:** `docs/` es del PROYECTO. MVGN despliega las plantillas
   iniciales ahí una vez; después son propiedad exclusiva del proyecto.
3. **Numeración de docs:** `NN_nombre.md`, dos dígitos, sin duplicados. `00–09` reservada.
4. **Publicación segura por construcción:** `.gitignore` ignora `.mvgn/` completo.
   Subir el proyecto a GitHub nunca filtra maquinaria ni runtime de MVGN.
5. **Todo documento lleva header DOC-REF** (`DOC-REF / SIG / DIVISION / NODE / STATE`
   + `ΚΛΕΙΣ`). Un documento sin `STATE` explícito no cumple este contrato.

   **Vocabulario de estado (bilingüe funcional):**

   | STATE | Griego | Significado |
   |:------|:-------|:------------|
   | `OPEN` | ΑΝΟΙΧΤΟ | Registro abierto, en captura |
   | `DRAFT` | ΠΡΟΧΕΙΡΟ | Borrador, sujeto a cambio |
   | `ACTIVE` | ΕΝΕΡΓΟ | Vivo, en mantenimiento |
   | `SEALED` | ΕΣΦΡΑΓΙΣΜΕΝΟ | Cerrado por AEON — inmutable |
   | `ARCHIVED` | ΑΡΧΕΙΟΘΕΤΗΜΕΝΟ | Retirado, preservado por referencia |

6. Los artefactos que se sellan pasan por AEON: solo se sello lo verificado.
7. Cambios a este contrato requieren bump de versión (`LAYOUT/x.y`) y entrada en changelog.

## 3. Qué NUNCA se publica de un proyecto con MVGN

Por construcción: `.mvgn/` completo está ignorado. Si el adoptante quiere trazabilidad
pública del blackboard, lo versiona conscientemente negando la ruta en su `.gitignore`.

---

*Crimson Nexus Corporation · MVGN Labs · Node MX-03*
