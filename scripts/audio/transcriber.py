#!/usr/bin/env python3
"""
MVGN Audio Transcriber — Faster-Whisper
Transcribe audio files (mp3, wav, m4a, flac, ogg) a texto.
Uso: python transcriber.py <audio_path> <output_dir> [model_size] [device]
"""

import sys
import os
import json
import time
from pathlib import Path
from datetime import datetime

# Salida en UTF-8 siempre (evita 'charmap' codec error con ✓/á/ñ en Windows)
if sys.stdout and hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
if sys.stderr and hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")


def get_device(device: str = "auto") -> str:
    """Detecta el dispositivo disponible."""
    if device != "auto":
        return device
    try:
        import torch
        return "cuda" if torch.cuda.is_available() else "cpu"
    except ImportError:
        return "cpu"


def transcribe(audio_path: str, output_dir: str, model_size: str = "small", device: str = "auto", engine: str = "faster-whisper", compute_type: str = "") -> dict:
    """
    Transcribe un archivo de audio.

    Args:
        audio_path: Ruta al archivo de audio
        output_dir: Directorio de salida para los artefactos
        model_size: Tamaño del modelo (tiny, base, small, medium, large-v3, turbo)
        device: Dispositivo (auto, cpu, cuda)
        engine: Motor de transcripción (faster-whisper, whisper)
        compute_type: Tipo de cómputo ("" = auto: float16 en cuda, int8 en cpu).
            Valores útiles en GPU con poca VRAM: "int8", "int8_float16"

    Returns:
        Dict con metadatos de la transcripción
    """
    device = get_device(device)
    audio_name = Path(audio_path).stem
    os.makedirs(output_dir, exist_ok=True)

    if engine == "faster-whisper":
        from faster_whisper import WhisperModel

        if not compute_type:
            compute_type = "float16" if device == "cuda" else "int8"
        cpu_threads = os.cpu_count() or 4

        print(f"[MVGN-AUDIO] Motor: Faster-Whisper | Modelo: {model_size} | Dispositivo: {device} ({compute_type})")
        start = time.time()

        model = WhisperModel(
            model_size,
            device=device,
            compute_type=compute_type,
            cpu_threads=cpu_threads,
            num_workers=2
        )

        print(f"[MVGN-AUDIO] Modelo cargado en {time.time() - start:.1f}s")

        print(f"[MVGN-AUDIO] Transcribiendo {audio_path}...")

        segments, info = model.transcribe(
            audio_path,
            beam_size=5,
            vad_filter=True,
            vad_parameters=dict(
                min_silence_duration_ms=500,
                threshold=0.5
            )
        )

        # Construir metadata
        all_segments = []
        full_text_parts = []
        total_duration = float(getattr(info, 'duration', 0) or 0)

        for segment in segments:
            seg_data = {
                "start": round(segment.start, 2),
                "end": round(segment.end, 2),
                "text": segment.text.strip()
            }
            all_segments.append(seg_data)
            full_text_parts.append(f"[{seg_data['start']:.2f}s -> {seg_data['end']:.2f}s] {seg_data['text']}")

            # Progreso en vivo: el script PowerShell lo convierte en barra (Write-Progress)
            if total_duration > 0:
                pct = min(100.0, seg_data["end"] / total_duration * 100.0)
                print(f"[MVGN-AUDIO][PROGRESS] {pct:.1f}|{seg_data['end']:.0f}|{total_duration:.0f}", flush=True)

        metadata = {
            "engine": "faster-whisper",
            "audio_file": audio_path,
            "audio_name": audio_name,
            "model": model_size,
            "device": device,
            "compute_type": compute_type,
            "language": info.language,
            "language_probability": round(info.language_probability, 3),
            "duration_seconds": round(info.duration, 2) if hasattr(info, 'duration') else 0,
            "transcribed_at": datetime.now().astimezone().isoformat(timespec="seconds"),
            "segments_count": len(all_segments),
            "total_text_length": len("\n\n".join(full_text_parts))
        }
        full_text = "\n\n".join(full_text_parts)

    elif engine == "whisper":
        import whisper

        print(f"[MVGN-AUDIO] Motor: OpenAI Whisper | Modelo: {model_size} | Dispositivo: {device}")
        start = time.time()

        model = whisper.load_model(model_size, device=device)

        print(f"[MVGN-AUDIO] Modelo cargado en {time.time() - start:.1f}s")
        print(f"[MVGN-AUDIO] Transcribiendo {audio_path}...")

        result = model.transcribe(audio_path, language=None, verbose=False)

        all_segments = []
        full_text_parts = []
        total_duration = float(all_segments[-1]["end"]) if all_segments else 0
        # El loop de OpenAI Whisper no expone progreso incremental
        # (solo la duración total); se reporta avance por segmento cuando hay más de 1
        seg_index = 0
        for seg in result.get("segments", []):
            seg_data = {
                "start": round(seg.get("start", 0), 2),
                "end": round(seg.get("end", 0), 2),
                "text": seg.get("text", "").strip()
            }
            all_segments.append(seg_data)
            full_text_parts.append(f"[{seg_data['start']:.2f}s -> {seg_data['end']:.2f}s] {seg_data['text']}")

            total_duration = float(all_segments[-1]["end"]) if all_segments else 0
            if total_duration > 0:
                pct = min(100.0, seg_data["end"] / total_duration * 100.0)
                print(f"[MVGN-AUDIO][PROGRESS] {pct:.1f}|{seg_data['end']:.0f}|{total_duration:.0f}", flush=True)
            seg_index += 1

        metadata = {
            "engine": "whisper",
            "audio_file": audio_path,
            "audio_name": audio_name,
            "model": model_size,
            "device": device,
            "language": result.get("language", "unknown"),
            "duration_seconds": round(all_segments[-1]["end"], 2) if all_segments else 0,
            "transcribed_at": datetime.now().astimezone().isoformat(timespec="seconds"),
            "segments_count": len(all_segments),
            "total_text_length": len("\n\n".join(full_text_parts))
        }

    else:
        raise ValueError(f"Motor desconocido: {engine}. Usa 'faster-whisper' o 'whisper'.")

    # Guardar transcripción completa
    transcript_path = os.path.join(output_dir, "transcript.md")
    with open(transcript_path, "w", encoding="utf-8") as f:
        f.write(f"# Transcripción de Audio\n\n")
        f.write(f"**Archivo:** {audio_path}\n")
        f.write(f"**Idioma:** {info.language} ({info.language_probability:.1%})\n")
        f.write(f"**Modelo:** {model_size} ({device}/{compute_type})\n")
        f.write(f"**Duración:** {metadata['duration_seconds']:.1f}s\n")
        f.write(f"**Segmentos:** {len(all_segments)}\n\n")
        f.write("---\n\n")
        f.write(full_text)
        f.write("\n")

    print(f"[MVGN-AUDIO] ✓ Transcripción guardada: {transcript_path}")
    print(f"[MVGN-AUDIO]   Idioma: {info.language} ({info.language_probability:.1%})")
    print(f"[MVGN-AUDIO]   Segmentos: {len(all_segments)}")
    print(f"[MVGN-AUDIO]   Duración: {metadata['duration_seconds']:.1f}s")

    # Guardar metadata
    meta_path = os.path.join(output_dir, "metadata.json")
    with open(meta_path, "w", encoding="utf-8") as f:
        json.dump(metadata, f, indent=2, ensure_ascii=False)

    # Guardar segmentos como JSON (para procesamiento)
    seg_path = os.path.join(output_dir, "segments.json")
    with open(seg_path, "w", encoding="utf-8") as f:
        json.dump(all_segments, f, indent=2, ensure_ascii=False)

    return metadata


def main():
    if len(sys.argv) < 3:
        print("Uso: python transcriber.py <audio_path> <output_dir> [model_size] [device] [engine] [compute_type]")
        print("  model_size: tiny, base, small (default), medium, large-v3, turbo")
        print("  device: auto (default), cpu, cuda")
        print("  engine: faster-whisper (default), whisper")
        print("  compute_type: '' (auto), int8, int8_float16, float16")
        sys.exit(1)

    audio_path = sys.argv[1]
    output_dir = sys.argv[2]
    model_size = sys.argv[3] if len(sys.argv) > 3 else "small"
    device = sys.argv[4] if len(sys.argv) > 4 else "auto"
    engine = sys.argv[5] if len(sys.argv) > 5 else "faster-whisper"
    compute_type = sys.argv[6] if len(sys.argv) > 6 else ""

    if not os.path.exists(audio_path):
        print(f"ERROR: Archivo no encontrado: {audio_path}")
        sys.exit(1)

    # Validar extensión
    valid_extensions = {".mp3", ".wav", ".m4a", ".flac", ".ogg"}
    ext = os.path.splitext(audio_path)[1].lower()
    if ext not in valid_extensions:
        print(f"ERROR: Extensión no soportada '{ext}'. Válidas: {', '.join(valid_extensions)}")
        sys.exit(1)

    try:
        metadata = transcribe(audio_path, output_dir, model_size, device, engine, compute_type)
        print(f"\n[MVGN-AUDIO] Transcripción completada exitosamente.")
    except ImportError as e:
        print(f"\nERROR: Dependencia faltante: {e}")
        print("")
        if engine == "faster-whisper":
            print("Instala las dependencias:")
            print("  pip install faster-whisper")
        else:
            print("Instala las dependencias:")
            print("  pip install openai-whisper")
        sys.exit(2)
    except Exception as e:
        print(f"\nERROR durante transcripción: {e}")
        sys.exit(3)


if __name__ == "__main__":
    main()
