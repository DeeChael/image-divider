package main

import (
	"bufio"
	"flag"
	"go.uber.org/zap"
	image2 "image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	logger := zap.Must(zap.NewProduction())
	zap.ReplaceGlobals(logger)

	var imageFile string
	var clips int
	var outputDirectory string
	var outputPrefix string
	var format string

	var isInCurrentDirectory bool = true

	var outputExactPrefix string

	flag.StringVar(&imageFile, "i", "", "Image file")
	flag.StringVar(&imageFile, "input", "", "Image file")
	flag.IntVar(&clips, "c", -1, "How many clips will the image be divided")
	flag.IntVar(&clips, "clips", -1, "How many clips will the image be divided")
	flag.StringVar(&outputDirectory, "o", "", "Output Directory")
	flag.StringVar(&outputDirectory, "output", "", "Output Directory")
	flag.StringVar(&outputPrefix, "p", "", "Prefix of the output files")
	flag.StringVar(&outputPrefix, "prefix", "", "Prefix of the output files")
	flag.StringVar(&format, "f", "png", "The format of output image files")
	flag.StringVar(&format, "format", "png", "The format of output image files")
	flag.Parse()

	if imageFile == "" {
		flag.PrintDefaults()
		return
	}

	if clips == -1 {
		flag.PrintDefaults()
		return
	}

	_, err := os.Stat(imageFile)

	if os.IsNotExist(err) {
		logger.Fatal("The image file does not exist")
		return
	} else if err != nil {
		logger.Fatal("Cannot get the info of the image file")
		return
	}

	if format != "png" && format != "jpg" {
		logger.Fatal("Unsupported output file format!")
		return
	}

	if outputDirectory != "" {
		isInCurrentDirectory = false
	}

	if outputPrefix == "" {
		outputExactPrefix = imageFile + "_output_"
	} else {
		if strings.HasSuffix(outputPrefix, "_") {
			outputExactPrefix = outputPrefix
		} else {
			outputExactPrefix = outputPrefix + "_"
		}
	}

	if !isInCurrentDirectory {
		if !strings.HasSuffix(outputDirectory, "/") {
			outputExactPrefix = "/" + outputExactPrefix
		}
		outputExactPrefix = outputDirectory + outputExactPrefix
	}

	_, err = os.Stat(outputDirectory)
	if os.IsNotExist(err) {
		logger.Fatal("The output directory does not exist")
		return
	} else if err != nil {
		logger.Fatal("Cannot get the info of the output directory")
		return
	}

	if clips < 1 {
		logger.Fatal("The image cannot be divided into a number of clips smaller than 1")
	} else if clips == 1 {
		file, err := os.Create(outputExactPrefix + "0." + format)
		if err != nil {
			logger.Fatal(err.Error())
			return
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				logger.Fatal("Error occurs when trying to open the output file")
			}
		}(file)

		input, _ := os.Open(imageFile)
		defer func(input *os.File) {
			err := input.Close()
			if err != nil {
				logger.Fatal("Error occurs when closing the image file")
			}
		}(input)

		reader := bufio.NewReader(input)
		writer := bufio.NewWriter(file)

		_, err = io.Copy(writer, reader)
		if err != nil {
			logger.Fatal("Error occurs when writing data to output file")
			return
		}

		logger.Info("Divided image successfully")
	}

	input, _ := os.Open(imageFile)
	defer func(input *os.File) {
		err := input.Close()
		if err != nil {
			logger.Fatal("Error occurs when closing the image file")
		}
	}(input)

	image, _, err := image2.Decode(input)
	if err != nil {
		logger.Fatal("Error occurs when decoding the image file")
		return
	}

	gap := image.Bounds().Size().Y / clips

	gapLimit := image.Bounds().Size().Y / (clips - 1)

	for image.Bounds().Size().Y%gap != 0 && gap < gapLimit && gap*clips < image.Bounds().Size().Y {
		gap++
	}

	height := 0

	for i := 0; i < clips; i++ {
		rect := image2.Rect(0, 0, image.Bounds().Size().X, Min(gap, image.Bounds().Size().Y-height))
		newRGBA := image2.NewRGBA(rect)

		draw.Draw(newRGBA, rect, image, image2.Point{Y: height}, draw.Src)

		newImage := newRGBA.SubImage(rect)

		if format == "img" {
			saveJpg(logger, newImage, outputExactPrefix+strconv.Itoa(i)+"."+format)
		} else {
			savePng(logger, newImage, outputExactPrefix+strconv.Itoa(i)+"."+format)
		}

		height += gap
	}

	logger.Info("Successfully divided the image!")

}

func saveJpg(logger *zap.Logger, image image2.Image, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		logger.Fatal(err.Error())
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Fatal("Error occurs when closing the output image file")
		}
	}(file)

	options := &jpeg.Options{Quality: 100}
	err = jpeg.Encode(file, image, options)
	if err != nil {
		logger.Fatal("Error occurs when encoding the image")
		return
	}
}

func savePng(logger *zap.Logger, image image2.Image, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		logger.Fatal(err.Error())
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Fatal("Error occurs when closing the output image file")
		}
	}(file)

	err = png.Encode(file, image)
	if err != nil {
		logger.Fatal("Error occurs when encoding the image")
		return
	}
}

func Min(n1 int, n2 int) int {
	if n1 < n2 {
		return n1
	}
	return n2
}
