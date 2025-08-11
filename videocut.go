package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Segment struct {
	Start int
	End   int
}

type VideoConfig struct {
	InputFile   string
	OutputFile  string
	Segments    []Segment
	CRF         string
	Preset      string
	Width       string
	Height      string
	FPS         string
	HWAccel     string
	Threads     string
	TwoPass     bool
	OptimizeFor string
}

func parseTime(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("tiempo vacío")
	}

	if strings.Contains(s, ":") {
		parts := strings.Split(s, ":")
		switch len(parts) {
		case 2:
			mm, errM := strconv.Atoi(parts[0])
			ss, errS := strconv.Atoi(parts[1])
			if errM != nil || errS != nil || mm < 0 || ss < 0 || ss > 59 {
				return 0, fmt.Errorf("formato inválido: %q", s)
			}
			return mm*60 + ss, nil
		case 3:
			hh, errH := strconv.Atoi(parts[0])
			mm, errM := strconv.Atoi(parts[1])
			ss, errS := strconv.Atoi(parts[2])
			if errH != nil || errM != nil || errS != nil || hh < 0 || mm < 0 || mm > 59 || ss < 0 || ss > 59 {
				return 0, fmt.Errorf("formato inválido: %q", s)
			}
			return hh*3600 + mm*60 + ss, nil
		default:
			return 0, fmt.Errorf("formato inválido: %q", s)
		}
	}

	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return 0, fmt.Errorf("no puedo parsear tiempo: %q", s)
	}
	return v, nil
}

func parseSegmentLine(line string) (Segment, error) {
	reSpace := regexp.MustCompile(`[,\s]+`)
	parts := reSpace.Split(strings.TrimSpace(line), -1)

	if len(parts) < 2 {
		return Segment{}, errors.New("faltan tiempos (formato: inicio fin)")
	}

	start, err1 := parseTime(parts[0])
	end, err2 := parseTime(parts[1])

	if err1 != nil {
		return Segment{}, fmt.Errorf("tiempo de inicio inválido: %v", err1)
	}
	if err2 != nil {
		return Segment{}, fmt.Errorf("tiempo de fin inválido: %v", err2)
	}
	if end <= start {
		return Segment{}, errors.New("el tiempo de fin debe ser mayor al de inicio")
	}

	return Segment{Start: start, End: end}, nil
}

func loadSegmentsFromFile(tsFile string) ([]Segment, error) {
	f, err := os.Open(tsFile)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir archivo de timestamps: %v", err)
	}
	defer f.Close()

	var segments []Segment
	sc := bufio.NewScanner(f)
	lineNum := 0

	for sc.Scan() {
		lineNum++
		line := strings.TrimSpace(sc.Text())

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		segment, err := parseSegmentLine(line)
		if err != nil {
			return nil, fmt.Errorf("línea %d: %v", lineNum, err)
		}
		segments = append(segments, segment)
	}

	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("error leyendo archivo: %v", err)
	}

	if len(segments) == 0 {
		return nil, errors.New("no se encontraron segmentos válidos")
	}

	return segments, nil
}

func loadMultiConfig(configFile string) ([]VideoConfig, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir archivo de configuración: %v", err)
	}
	defer f.Close()

	var configs []VideoConfig
	var currentConfig *VideoConfig
	sc := bufio.NewScanner(f)
	lineNum := 0

	for sc.Scan() {
		lineNum++
		line := strings.TrimSpace(sc.Text())

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if currentConfig != nil {
				if err := validateConfig(currentConfig); err != nil {
					return nil, fmt.Errorf("línea %d: configuración anterior inválida: %v", lineNum, err)
				}
				configs = append(configs, *currentConfig)
			}

			currentConfig = &VideoConfig{
				CRF:         "20",
				Preset:      "veryfast",
				Width:       "1920",
				Height:      "1080",
				FPS:         "30",
				HWAccel:     "auto",
				Threads:     "0",
				TwoPass:     false,
				OptimizeFor: "balanced",
			}
			continue
		}

		if currentConfig == nil {
			return nil, fmt.Errorf("línea %d: debe comenzar con una sección [nombre]", lineNum)
		}

		if strings.Contains(line, "=") {
			if err := parseConfigLine(line, currentConfig); err != nil {
				return nil, fmt.Errorf("línea %d: %v", lineNum, err)
			}
		} else {
			segment, err := parseSegmentLine(line)
			if err != nil {
				return nil, fmt.Errorf("línea %d: %v", lineNum, err)
			}
			currentConfig.Segments = append(currentConfig.Segments, segment)
		}
	}

	if currentConfig != nil {
		if err := validateConfig(currentConfig); err != nil {
			return nil, fmt.Errorf("configuración final inválida: %v", err)
		}
		configs = append(configs, *currentConfig)
	}

	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("error leyendo archivo: %v", err)
	}

	if len(configs) == 0 {
		return nil, errors.New("no se encontraron configuraciones de video")
	}

	return configs, nil
}

func parseConfigLine(line string, config *VideoConfig) error {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("formato inválido, esperado clave=valor")
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch key {
	case "input":
		config.InputFile = value
	case "output":
		config.OutputFile = value
	case "crf":
		config.CRF = value
	case "preset":
		config.Preset = value
	case "width":
		config.Width = value
	case "height":
		config.Height = value
	case "fps":
		config.FPS = value
	case "hwaccel":
		config.HWAccel = value
	case "threads":
		config.Threads = value
	case "twopass":
		config.TwoPass = strings.ToLower(value) == "true" || value == "1"
	case "optimize":
		config.OptimizeFor = value
	default:
		return fmt.Errorf("parámetro desconocido: %s", key)
	}
	return nil
}

func validateConfig(config *VideoConfig) error {
	if config.InputFile == "" {
		return errors.New("falta archivo de entrada (input)")
	}
	if config.OutputFile == "" {
		return errors.New("falta archivo de salida (output)")
	}

	if len(config.Segments) == 0 {
		return errors.New("no hay segmentos definidos")
	}
	return nil
}

func detectOptimalEncoder() (string, map[string]string) {
	encoders := []struct {
		hwaccel string
		codec   string
		testCmd []string
		info    map[string]string
	}{
		{
			hwaccel: "nvenc",
			codec:   "h264_nvenc",
			testCmd: []string{"ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc2=duration=0.1:size=320x240:rate=1", "-c:v", "h264_nvenc", "-preset", "fast", "-t", "0.1", "-f", "null", "-"},
			info:    map[string]string{"codec": "h264_nvenc", "type": "NVIDIA GPU", "performance": "High"},
		},
		{
			hwaccel: "qsv",
			codec:   "h264_qsv",
			testCmd: []string{"ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc2=duration=0.1:size=320x240:rate=1", "-c:v", "h264_qsv", "-preset", "fast", "-t", "0.1", "-f", "null", "-"},
			info:    map[string]string{"codec": "h264_qsv", "type": "Intel QuickSync", "performance": "Medium"},
		},
		{
			hwaccel: "vaapi",
			codec:   "h264_vaapi",
			testCmd: []string{"ffmpeg", "-hide_banner", "-loglevel", "error", "-init_hw_device", "vaapi=foo:/dev/dri/renderD128", "-f", "lavfi", "-i", "testsrc2=duration=0.1:size=320x240:rate=1", "-vf", "format=nv12,hwupload", "-c:v", "h264_vaapi", "-t", "0.1", "-f", "null", "-"},
			info:    map[string]string{"codec": "h264_vaapi", "type": "Intel/AMD VAAPI", "performance": "Medium"},
		},
	}

	fmt.Println("🔍 Detectando hardware de aceleración disponible...")

	for _, encoder := range encoders {
		fmt.Printf("   Probando %s (%s)... ", encoder.info["type"], encoder.codec)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, encoder.testCmd[0], encoder.testCmd[1:]...)
		cmd.Stdout = nil
		cmd.Stderr = nil

		if err := cmd.Run(); err == nil {
			fmt.Printf("✅ Disponible\n")
			cancel()
			return encoder.hwaccel, encoder.info
		} else {
			fmt.Printf("❌ No disponible\n")
		}
		cancel()
	}

	fmt.Println("   🔄 Usando libx264 (CPU) como respaldo")
	return "cpu", map[string]string{"codec": "libx264", "type": "CPU", "performance": "Variable"}
}

func buildFFmpegCommand(config VideoConfig) ([]string, error) {
	args := []string{"-hide_banner", "-y"}

	hwAccel := config.HWAccel
	var encoderConfig map[string]string
	if hwAccel == "" || hwAccel == "auto" {
		hwAccel, encoderConfig = detectOptimalEncoder()
		fmt.Printf("🔧 Encoder detectado: %s (%s)\n", hwAccel, encoderConfig["codec"])
	}

	if hwAccel == "nvenc" {
		args = append(args, "-hwaccel", "cuda", "-hwaccel_output_format", "cuda")
		args = append(args, "-avoid_negative_ts", "make_zero")
	} else if hwAccel == "qsv" {
		args = append(args, "-hwaccel", "qsv")
	} else if hwAccel == "vaapi" {
		args = append(args, "-hwaccel", "vaapi", "-hwaccel_device", "/dev/dri/renderD128")
	}

	if config.Threads != "" {
		args = append(args, "-threads", config.Threads)
	} else {
		args = append(args, "-threads", "0")
	}

	args = append(args, "-i", config.InputFile)

	var filters []string
	var inputs []string

	for i, segment := range config.Segments {
		filters = append(filters,
			fmt.Sprintf("[0:v]scale=%s:%s:force_original_aspect_ratio=decrease,pad=%s:%s:(ow-iw)/2:(oh-ih)/2,fps=%s,trim=start=%d:end=%d,setpts=PTS-STARTPTS[v%d]",
				config.Width, config.Height, config.Width, config.Height, config.FPS, segment.Start, segment.End, i),
			fmt.Sprintf("[0:a]atrim=start=%d:end=%d,asetpts=PTS-STARTPTS[a%d]", segment.Start, segment.End, i),
		)
		inputs = append(inputs, fmt.Sprintf("[v%d][a%d]", i, i))
	}

	n := len(inputs)
	filters = append(filters, strings.Join(inputs, "")+fmt.Sprintf("concat=n=%d:v=1:a=1[v][a]", n))

	var videoCodec, videoPreset string
	if hwAccel == "nvenc" {
		videoCodec = "h264_nvenc"
		videoPreset = mapPresetToNvenc(config.Preset)
		args = append(args, "-rc", "vbr", "-cq", config.CRF)
	} else if hwAccel == "qsv" {
		videoCodec = "h264_qsv"
		videoPreset = "balanced"
		args = append(args, "-global_quality", config.CRF)
	} else if hwAccel == "vaapi" {
		videoCodec = "h264_vaapi"
		videoPreset = "fast"
		args = append(args, "-qp", config.CRF)
	} else {
		videoCodec = "libx264"
		videoPreset = config.Preset
	}

	args = append(args,
		"-filter_complex", strings.Join(filters, ";"),
		"-map", "[v]", "-map", "[a]",
		"-c:v", videoCodec,
	)

	if videoCodec == "libx264" {
		args = append(args, "-crf", config.CRF, "-preset", videoPreset)
		args = append(args, "-x264-params", "ref=1:bframes=1:me=hex:subme=1")
	} else {
		args = append(args, "-preset", videoPreset)
	}

	args = append(args,
		"-c:a", "aac", "-b:a", "192k",
		"-movflags", "+faststart",
		config.OutputFile,
	)

	return args, nil
}

func mapPresetToNvenc(preset string) string {
	presetMap := map[string]string{
		"ultrafast": "p1",
		"superfast": "p2",
		"veryfast":  "p3",
		"faster":    "p4",
		"fast":      "p5",
		"medium":    "p6",
		"slow":      "p7",
		"slower":    "p7",
		"veryslow":  "p7",
	}
	if nvencPreset, exists := presetMap[preset]; exists {
		return nvencPreset
	}
	return "p4"
}

func processVideo(config VideoConfig, videoNum, totalVideos int) error {
	fmt.Printf("\n🎬 [%d/%d] Procesando: %s -> %s\n", videoNum, totalVideos, config.InputFile, config.OutputFile)
	fmt.Printf("🎯 Config: %sx%s@%sfps, CRF:%s, Preset:%s, Segmentos:%d\n",
		config.Width, config.Height, config.FPS, config.CRF, config.Preset, len(config.Segments))

	ffmpegArgs, err := buildFFmpegCommand(config)
	if err != nil {
		return fmt.Errorf("error construyendo comando: %v", err)
	}

	fmt.Println(">> ffmpeg", strings.Join(ffmpegArgs, " "))

	cmd := exec.Command("ffmpeg", ffmpegArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error ejecutando ffmpeg: %v", err)
	}

	fmt.Printf("✅ Completado: %s\n", config.OutputFile)
	return nil
}

func printUsage() {
	fmt.Println(`VideoCut - Editor de video simplificado con configuración unificada

USO:
  go run videocut.go config.txt

El archivo de configuración debe contener una o más secciones con este formato:

[nombre_configuracion]
input = video_entrada.mp4
output = video_salida.mp4
width = 1920
height = 1080
fps = 30
crf = 22
preset = fast
hwaccel = auto
# Timestamps directamente en la configuración (formato: start end)
00:01:30 00:02:45
00:05:00 00:07:30

Opciones disponibles:
- input: Archivo de video de entrada (requerido)
- output: Archivo de video de salida (requerido) 
- width: Ancho del video (defecto: 1920)
- height: Alto del video (defecto: 1080)
- fps: Cuadros por segundo (defecto: 30)
- crf: Calidad del video 18-28 (defecto: 22)
- preset: Velocidad encoding ultrafast|veryfast|fast|medium|slow (defecto: fast)
- hwaccel: Aceleración hardware auto|nvenc|qsv|vaapi|cpu (defecto: auto)
- threads: Número de threads (defecto: 0=auto)
- twopass: true para encoding de dos pasadas
- optimize: speed|balanced|quality (defecto: balanced)

Hardware Acceleration:
- nvenc: NVIDIA GPUs (GeForce/Quadro/Tesla)
- qsv: Intel Quick Sync Video  
- vaapi: Intel/AMD GPUs (Linux)
- cpu: Siempre funciona, compatible con todas las funciones
- auto: Detecta automáticamente la mejor opción`)
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 || len(args) != 1 {
		printUsage()
		os.Exit(2)
	}

	configFile := args[0]
	configs, err := loadMultiConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error cargando configuración: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📋 Configuración cargada: %d video(s) para procesar\n", len(configs))

	successCount := 0
	for i, config := range configs {
		if err := processVideo(config, i+1, len(configs)); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error procesando %s: %v\n", config.InputFile, err)
			continue
		}
		successCount++
	}

	fmt.Printf("\n🎉 Proceso completado: %d/%d videos procesados exitosamente\n", successCount, len(configs))
	if successCount < len(configs) {
		os.Exit(1)
	}
}
