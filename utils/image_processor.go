package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/image/webp"
)

// ImageProcessor обрабатывает изображения товаров
type ImageProcessor struct {
	TargetWidth     int
	TargetHeight    int
	BackgroundColor color.Color
	JPEGQuality     int
}

// NewImageProcessor создает новый процессор изображений
func NewImageProcessor(width, height int, bgColor string) *ImageProcessor {
	processor := &ImageProcessor{
		TargetWidth:  width,
		TargetHeight: height,
		JPEGQuality:  85, // Хороший баланс между качеством и размером
	}

	// Парсим цвет фона
	switch strings.ToLower(bgColor) {
	case "white":
		processor.BackgroundColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	case "transparent":
		processor.BackgroundColor = color.Transparent
	default:
		processor.BackgroundColor = color.RGBA{R: 255, G: 255, B: 255, A: 255} // По умолчанию белый
	}

	return processor
}

// ProcessProductImage обрабатывает изображение товара:
// 1. Читает и декодирует изображение с учетом EXIF ориентации (для фото с телефонов)
// 2. Изменяет размер с сохранением пропорций
// 3. Добавляет фон (белый/прозрачный)
// 4. Центрирует изображение
// 5. Сжимает в JPEG
func (ip *ImageProcessor) ProcessProductImage(input io.Reader, outputPath string) (int64, error) {
	// Читаем все данные в память (нужно для обработки EXIF)
	data, err := io.ReadAll(input)
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения файла: %v", err)
	}

	// Декодируем изображение с автоматической обработкой EXIF ориентации
	// imaging.Decode автоматически поворачивает изображение согласно EXIF данным
	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("ошибка декодирования изображения: %v", err)
	}

	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()
	log.Printf("📸 Обработка изображения с телефона: размер=%dx%d (после обработки EXIF ориентации)", originalWidth, originalHeight)

	// Создаем новое изображение с нужным размером и фоном
	bg := image.NewRGBA(image.Rect(0, 0, ip.TargetWidth, ip.TargetHeight))

	// Заливаем фон (если не прозрачный)
	if ip.BackgroundColor != color.Transparent {
		draw.Draw(bg, bg.Bounds(), &image.Uniform{ip.BackgroundColor}, image.Point{}, draw.Src)
	}

	// Изменяем размер исходного изображения с сохранением пропорций
	// Используем библиотеку imaging для качественного масштабирования
	resized := imaging.Fit(img, ip.TargetWidth, ip.TargetHeight, imaging.Lanczos)

	// Вычисляем позицию для центрирования
	bounds := resized.Bounds()
	x := (ip.TargetWidth - bounds.Dx()) / 2
	y := (ip.TargetHeight - bounds.Dy()) / 2

	// Рисуем изображение поверх фона (центрируем)
	draw.Draw(bg,
		image.Rect(x, y, x+bounds.Dx(), y+bounds.Dy()),
		resized,
		bounds.Min,
		draw.Over)

	// Создаем директорию, если её нет
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("ошибка создания директории: %v", err)
	}

	// Сохраняем результат
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания файла: %v", err)
	}
	defer outputFile.Close()

	// Сохраняем как JPEG (всегда, для единообразия и лучшего сжатия)
	err = jpeg.Encode(outputFile, bg, &jpeg.Options{Quality: ip.JPEGQuality})
	if err != nil {
		return 0, fmt.Errorf("ошибка кодирования JPEG: %v", err)
	}

	// Получаем размер файла
	info, err := outputFile.Stat()
	if err != nil {
		return 0, err
	}

	log.Printf("✅ Изображение обработано: %s, размер=%d байт", outputPath, info.Size())
	return info.Size(), nil
}

// DecodeImage декодирует изображение из разных форматов
// Примечание: для обработки фото товаров используйте ProcessProductImage,
// который автоматически обрабатывает EXIF ориентацию (важно для фото с телефонов)
func DecodeImage(r io.Reader, contentType string) (image.Image, string, error) {
	// Пробуем определить формат по Content-Type
	switch {
	case strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg"):
		img, err := jpeg.Decode(r)
		return img, "jpeg", err
	case strings.Contains(contentType, "png"):
		img, err := png.Decode(r)
		return img, "png", err
	case strings.Contains(contentType, "webp"):
		img, err := webp.Decode(r)
		return img, "webp", err
	default:
		// Пробуем автоматически определить формат
		img, format, err := image.Decode(r)
		return img, format, err
	}
}

