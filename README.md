# VideoCut - Script Simplificado de Edición de Video

Script Go simplificado para crear highlights de video con timestamps integrados. Versión optimizada sin funciones de cover/end para máxima compatibilidad.

## ✨ Características Principales

- **🎯 Configuración Unificada**: Solo un archivo de configuración con timestamps integrados
- **⚡ Hardware Acceleration**: Detección automática de NVENC, QSV, VAAPI y CPU
- **🔧 Múltiples Perfiles**: 9 configuraciones predefinidas para diferentes casos de uso
- **📱 Multi-resolución**: Desde 720p móvil hasta 4K
- **🎮 Gaming Ready**: Configuraciones específicas para clips de gaming a 60fps

## 🚀 Instalación y Uso

### Prerrequisitos
- Go 1.19+ instalado
- FFmpeg con soporte de hardware acceleration (opcional)

### Uso Simple
```bash
go run videocut.go config.txt
```

## 📝 Configuración

El archivo `config.txt` contiene múltiples perfiles. Ejemplo:

```ini
[basico]
input = input.mp4
output = output_basico.mp4
hwaccel = auto
width = 1920
height = 1080
fps = 30
crf = 23
preset = fast
00:00:00 00:00:30
00:02:00 00:03:00

[gaming_60fps]
input = input.mp4
output = output_gaming.mp4
hwaccel = auto
width = 1920
height = 1080
fps = 60
crf = 22
preset = fast
00:00:00 00:00:30
```

### Parámetros Disponibles

| Parámetro  | Descripción          | Valores                                | Defecto       |
| ---------- | -------------------- | -------------------------------------- | ------------- |
| `input`    | Archivo de entrada   | Archivo MP4                            | **Requerido** |
| `output`   | Archivo de salida    | Archivo MP4                            | **Requerido** |
| `hwaccel`  | Aceleración hardware | `auto`, `nvenc`, `qsv`, `vaapi`, `cpu` | `auto`        |
| `width`    | Ancho del video      | Píxeles                                | `1920`        |
| `height`   | Alto del video       | Píxeles                                | `1080`        |
| `fps`      | Cuadros por segundo  | 24, 30, 60                             | `30`          |
| `crf`      | Calidad del video    | 18-28 (menor=mejor)                    | `22`          |
| `preset`   | Velocidad encoding   | `ultrafast`, `fast`, `medium`, `slow`  | `fast`        |
| `threads`  | Número de threads    | Número o `0` (auto)                    | `0`           |
| `twopass`  | Encoding 2 pasadas   | `true`, `false`                        | `false`       |
| `optimize` | Optimización         | `speed`, `balanced`, `quality`         | `balanced`    |

### Timestamps
Los timestamps se especifican directamente en el archivo de configuración:
```
HH:MM:SS HH:MM:SS
00:01:30 00:02:45
00:05:00 00:07:30
```

## 🎯 Perfiles Predefinidos

### Básicos
- **`basico`**: Configuración estándar 1080p
- **`rapido`**: Pruebas rápidas 720p con calidad reducida

### Calidad
- **`alta_calidad`**: Máxima calidad con encoding de 2 pasadas
- **`calidad_4k`**: Contenido 4K (requiere CPU potente)

### Especializados
- **`gaming_60fps`**: Clips de gaming a 60fps
- **`mobile_optimized`**: Optimizado para dispositivos móviles
- **`streaming_optimized`**: Balance para streaming/web

### Hardware Específico
- **`nvenc_test`**: Prueba específica con NVENC
- **`cpu_forzado`**: Forzar CPU para máxima compatibilidad

## ⚙️ Hardware Acceleration

### Compatibilidad
| Hardware      | Codec      | Notas                                     |
| ------------- | ---------- | ----------------------------------------- |
| **CPU**       | libx264    | ✅ Siempre funciona, máxima compatibilidad |
| **NVIDIA**    | h264_nvenc | ⚠️ Funciona solo para procesamiento simple |
| **Intel**     | h264_qsv   | ⚠️ Funciona solo para procesamiento simple |
| **AMD/Intel** | h264_vaapi | ⚠️ Linux principalmente                    |

### Recomendaciones
- **Auto-detección**: Usa `hwaccel = auto` para detección automática
- **Máxima compatibilidad**: Usa `hwaccel = cpu` si tienes problemas
- **Rendimiento**: NVENC/QSV son más rápidos pero menos compatibles con filtros complejos

## 📊 Resoluciones Comunes

| Resolución            | Uso Recomendado               | Configuración               |
| --------------------- | ----------------------------- | --------------------------- |
| **4K** (3840x2160)    | Contenido premium             | Requiere CPU potente        |
| **1440p** (2560x1440) | Gaming/streaming alta calidad | Balance calidad/rendimiento |
| **1080p** (1920x1080) | Estándar web/YouTube          | Recomendado general         |
| **720p** (1280x720)   | Móvil/pruebas rápidas         | Menor tamaño archivo        |

## 🎚️ Guía de Calidad (CRF)

| CRF       | Calidad   | Tamaño  | Uso                     |
| --------- | --------- | ------- | ----------------------- |
| **18-20** | Muy alta  | Grande  | Contenido final premium |
| **22-24** | Buena     | Medio   | Balance recomendado     |
| **26-28** | Aceptable | Pequeño | Pruebas/mobile          |

## 🔧 Resolución de Problemas

### NVENC Falla
**Problema**: Error "Impossible to convert between formats"  
**Solución**: Usar `hwaccel = cpu` en la configuración

### Videos No Procesados
**Problema**: No se generan archivos de salida  
**Verificar**:
1. Archivo de entrada existe
2. Timestamps son válidos
3. FFmpeg está instalado

### Rendimiento Lento
**Optimizar**:
1. Usar `preset = ultrafast` para velocidad
2. Reducir resolución a 720p
3. Usar hardware acceleration si funciona

## 📁 Estructura del Proyecto

```
cutcat_scripts/
├── videocut.go          # Script principal
├── config.txt           # Configuraciones predefinidas
├── input.mp4           # Video de entrada (ejemplo)
└── README.md           # Esta documentación
```

## 🔄 Changelog

### v2.0.0 - Simplificación Mayor
- ❌ **ELIMINADO**: Funciones cover/end (imágenes de portada/cierre)
- ❌ **ELIMINADO**: Modo simple con timestamps externos
- ❌ **ELIMINADO**: Soporte para archivos timestamps.txt separados
- ✅ **AÑADIDO**: Configuración unificada con timestamps integrados
- ✅ **MEJORADO**: Compatibilidad de hardware acceleration
- ✅ **SIMPLIFICADO**: Solo un modo de operación (config.txt)

### Cambios Técnicos
- Eliminadas funciones: `parseMediaSpec()`, `loadTimestamps()`, `parseSimpleArgs()`
- Eliminados campos: `CoverSpec`, `EndSpec`, `TimestampsFile`
- Simplificada función `buildFFmpegCommand()` sin filtros complejos de cover/end
- Mejorada detección automática de hardware

## 🤝 Contribuciones

Este es un proyecto simplificado enfocado en funcionalidad core. Para contribuir:

1. Fork el proyecto
2. Crear feature branch (`git checkout -b feature/mejora`)
3. Commit cambios (`git commit -am 'Añadir mejora'`)
4. Push branch (`git push origin feature/mejora`)
5. Crear Pull Request

## 📜 Licencia

Proyecto de código abierto. Libre para uso personal y comercial.