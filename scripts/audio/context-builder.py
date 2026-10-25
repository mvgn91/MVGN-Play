#!/usr/bin/env python3
"""
MVGN Context Builder — Genera context.md estructurado desde transcripción de audio.
Uso: python context-builder.py <transcript_dir> <output_dir>
"""

import sys
import os
import json
import re
from pathlib import Path

# Salida en UTF-8 siempre (evita 'charmap' codec error con ✓/á/ñ en Windows)
if sys.stdout and hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
if sys.stderr and hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")


def load_transcript(transcript_dir: str) -> tuple:
    """Carga la transcripción y metadatos."""
    transcript_path = os.path.join(transcript_dir, "transcript.md")
    meta_path = os.path.join(transcript_dir, "metadata.json")

    if not os.path.exists(transcript_path):
        print(f"ERROR: No se encuentra {transcript_path}")
        sys.exit(1)

    with open(transcript_path, "r", encoding="utf-8") as f:
        transcript = f.read()

    metadata = {}
    if os.path.exists(meta_path):
        with open(meta_path, "r", encoding="utf-8") as f:
            metadata = json.load(f)

    return transcript, metadata


def extract_entities(text: str) -> dict:
    """Extrae entidades, temas y conceptos del texto."""
    # Buscar palabras técnicas y nombres propios
    tech_pattern = r'\b([A-Z][a-z]+(?:\s[A-Z][a-z]+)*)\b'
    potential_entities = re.findall(tech_pattern, text)

    # Filtrar palabras comunes
    stop_words = {
        'Pero', 'Como', 'Cuando', 'Donde', 'Este', 'Esta', 'Esto', 'Eso',
        'Ese', 'Esa', 'Todo', 'Cada', 'Muy', 'Más', 'Menos', 'Antes',
        'Después', 'Sobre', 'Entre', 'Hasta', 'Desde', 'Durante',
        'Tiene', 'Hace', 'Hizo', 'Sido', 'Ser', 'Estar', 'Tener',
        'Puede', 'Debe', 'Vamos', 'Bueno', 'Entonces', 'Además',
        'Hello', 'Hi', 'Thank', 'Thanks', 'Yeah', 'Okay', 'Well'
    }
    entities = [e for e in potential_entities if e not in stop_words and len(e) > 2]
    # Palabras únicas, ordenadas por frecuencia
    from collections import Counter
    entity_counts = Counter(entities)
    return {
        "terminos_clave": [{"termino": term, "ocurrencias": count}
                          for term, count in entity_counts.most_common(30)]
    }


def extract_topics(text: str, sections: dict) -> dict:
    """Distribuye el texto en categorías temáticas."""
    # Buscar secciones del texto que coincidan con cada categoría
    results = {}
    for category, keywords in sections.items():
        lines = []
        for line in text.split('\n'):
            line_lower = line.lower()
            if any(kw.lower() in line_lower for kw in keywords):
                # Extraer la línea relevante
                clean_line = re.sub(r'\[\d+\.\d+s\s*->\s*\d+\.\d+s\]\s*', '', line).strip()
                if clean_line and len(clean_line) > 10:
                    lines.append(clean_line)
        results[category] = lines[:10]  # Top 10 líneas por categoría
    return results


def build_context(transcript: str, metadata: dict, output_dir: str):
    """Construye context.md estructurado."""
    # Secciones temáticas a detectar
    topic_sections = {
        "Objetivos Detectados": [
            "objetivo", "goal", "purpose", "queremos", "necesitamos",
            "buscamos", "propósito", "finalidad", "meta", "target"
        ],
        "Problemas Identificados": [
            "problema", "issue", "pain", "difícil", "complicado",
            "error", "falla", "bug", "limitación", "restricción"
        ],
        "Necesidades": [
            "necesita", "requiere", "need", "requirement", "debería",
            "hace falta", "carencia", "falta", "imprescindible"
        ],
        "Requisitos Funcionales": [
            "función", "feature", "funcionalidad", "capaz de",
            "debe poder", "permitir", "implementar", "crear"
        ],
        "Requisitos No Funcionales": [
            "rendimiento", "performance", "seguridad", "security",
            "escalable", "rápido", "rápida", "confiable"
        ],
        "Riesgos": [
            "riesgo", "risk", "peligro", "amenaza", "preocupación",
            "cuidado", "atención", "potencial problema"
        ],
        "Decisiones": [
            "decidimos", "vamos a usar", "elegimos", "optamos",
            "decisión", "acordamos", "concluimos"
        ],
        "Tecnologías Mencionadas": [
            "api", "framework", "librería", "library", "base de datos",
            "database", "servidor", "server", "python", "javascript",
            "typescript", "react", "node", "docker", "kubernetes",
            "aws", "azure", "gcp", "cloud", "microservicio"
        ],
        "Actores": [
            "usuario", "user", "cliente", "customer", "admin",
            "administrador", "manager", "coordinador"
        ],
        "Dudas / Preguntas Abiertas": [
            "duda", "pregunta", "question", "no sabemos",
            "habría que ver", "investigar", "uncertain"
        ],
        "Ideas Futuras": [
            "a futuro", "próximo paso", "next step", "futuro",
            "más adelante", "versión", "fase", "phase", "roadmap"
        ],
        "Acciones Pendientes": [
            "pendiente", "to-do", "todo", "por hacer",
            "falta hacer", "acción", "action item"
        ]
    }

    # Extraer entidades
    entities = extract_entities(transcript)

    # Extraer temas
    topics = extract_topics(transcript, topic_sections)

    # Metadata del audio
    audio_info = metadata.get("audio_file", "desconocido")
    language = metadata.get("language", "desconocido")
    duration = metadata.get("duration_seconds", 0)
    model = metadata.get("model", "desconocido")

    # Generar resumen ejecutivo automático
    lines = [line.strip() for line in transcript.split('\n')
             if line.strip() and not line.startswith('#') and not line.startswith('**')]
    clean_text = ' '.join(lines)
    # Primeros 500 chars como resumen
    summary = clean_text[:500] + "..." if len(clean_text) > 500 else clean_text

    # Construir context.md
    context = []
    context.append("# Contexto del Proyecto — Generado desde Audio\n")
    context.append("> Documento generado automáticamente a partir de transcripción de audio.\n")
    context.append("> Este archivo es la entrada oficial del generador de PRD.\n")
    context.append("")
    context.append("---\n")
    context.append("")

    # Resumen Ejecutivo
    context.append("## Resumen Ejecutivo\n")
    context.append(summary)
    context.append("")
    context.append("---\n")

    # Metadatos
    context.append("## Metadatos de la Fuente\n")
    context.append(f"| Campo | Valor |")
    context.append(f"|-------|-------|")
    context.append(f"| Archivo | `{audio_info}` |")
    context.append(f"| Idioma | {language} |")
    context.append(f"| Duración | {duration:.1f}s |")
    context.append(f"| Modelo | {model} |")
    context.append(f"| Generado | {metadata.get('transcribed_at', 'N/A')} |")
    context.append("")
    context.append("---\n")

    # Secciones temáticas
    section_count = 0
    for section_name, items in topics.items():
        if items:
            context.append(f"\n## {section_name}\n")
            for item in items:
                context.append(f"- {item}")
            context.append("")
            section_count += 1

    # Términos clave
    context.append("\n---\n")
    context.append("## Términos Clave Detectados\n")
    context.append("| Término | Ocurrencias |")
    context.append("|---------|-------------|")
    for entity in entities.get("terminos_clave", [])[:20]:
        context.append(f"| {entity['termino']} | {entity['ocurrencias']} |")
    context.append("")

    # Nota final
    context.append("\n---\n")
    context.append("*Contexto generado automáticamente por MVGN Audio Ingestion v1.0*\n")
    context.append("*Revisar y completar antes de generar el PRD.*\n")

    result = '\n'.join(context)

    # Guardar
    os.makedirs(output_dir, exist_ok=True)
    output_path = os.path.join(output_dir, "context.md")

    with open(output_path, "w", encoding="utf-8") as f:
        f.write(result)

    print(f"[MVGN-CONTEXT] ✓ Contexto generado: {output_path}")
    print(f"[MVGN-CONTEXT]   Secciones con contenido: {section_count}")
    print(f"[MVGN-CONTEXT]   Términos clave: {len(entities.get('terminos_clave', []))}")

    # Resumen para state.json
    summary_data = {
        "section_count": section_count,
        "total_terms": len(entities.get("terminos_clave", [])),
        "language": language,
        "duration_seconds": duration,
        "audio_file": audio_info
    }

    summary_path = os.path.join(output_dir, "context-summary.json")
    with open(summary_path, "w", encoding="utf-8") as f:
        json.dump(summary_data, f, indent=2, ensure_ascii=False)

    return result


def main():
    if len(sys.argv) < 2:
        print("Uso: python context-builder.py <transcript_dir> [output_dir]")
        print("  transcript_dir: Directorio donde están transcript.md y metadata.json")
        print("  output_dir:     Directorio de salida (default: transcript_dir)")
        sys.exit(1)

    transcript_dir = sys.argv[1]
    output_dir = sys.argv[2] if len(sys.argv) > 2 else transcript_dir

    transcript, metadata = load_transcript(transcript_dir)
    build_context(transcript, metadata, output_dir)

    print(f"\n[MVGN-CONTEXT] Contexto listo para generar PRD.")


if __name__ == "__main__":
    main()
