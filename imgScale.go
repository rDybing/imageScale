package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type imageT struct {
	width  int
	height int
}

func main() {
	var outFile [4]string
	inFile := [4]string{
		"test_iOS_24b.jpg",
		"test_iOS_24b_half.png",
		"test_iOS_32b.png",
		"test_iOS_32b_half.png"}
	inDir := "./imgPre/"
	outDir := "./imgPostScale/"
	cropDir := "./imgPostCrop/"

	for i := range inFile {
		outFile[i] = newSuffix(inFile[i])
		oldImg := getDimensions(inDir + inFile[i])
		fmt.Printf("old - w: %4d :: h: %4d\n", oldImg.width, oldImg.height)
		newImg := calcNewSize(oldImg)
		fmt.Printf("new - w: %4d :: h: %4d\n", newImg.width, newImg.height)
		scaleNewFile(inDir+inFile[i], outDir+outFile[i], newImg)
		cropNewFile(inDir+inFile[i], cropDir+outFile[i], newImg)
	}
}

func newSuffix(in string) string {
	if strings.HasSuffix(in, ".png") {
		in = strings.TrimSuffix(in, ".png")
	}
	if strings.HasSuffix(in, ".jpg") {
		in = strings.TrimSuffix(in, ".jpg")
	}
	in += ".small.png"
	return in
}

func getDimensions(inFile string) imageT {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "stream=width,height", "-of", "default=noprint_wrappers=1", inFile)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error running ffprobe: %v\n", err)
	}
	tStr := string(cmdOutput.Bytes())
	return cleanString(tStr)
}

func cropNewFile(inFile string, outFile string, img imageT) {
	height := strconv.Itoa(img.height)
	scaleCmd := `scale=(iw*sar)*max(` + height + `/(iw*sar)\,` + height + `/ih):ih*max(` + height + `/(iw*sar)\,` + height + `/ih), crop=256:256`
	cmd := exec.Command("ffmpeg", "-i", inFile, "-vf", scaleCmd, outFile)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error running ffmpeg (scale and crop): %v\n%s\n", err, stderr.String())
	}
	fmt.Println("Success Scale and Crop!")
}

func scaleNewFile(inFile string, outFile string, img imageT) {
	height := strconv.Itoa(img.height)
	width := strconv.Itoa(img.width)
	scaleCmd := `scale=(iw*sar)*max(` + width + `/(iw*sar)\,` + height + `/ih):ih*max(` + width + `/(iw*sar)\,` + height + `/ih)`
	cmd := exec.Command("ffmpeg", "-i", inFile, "-vf", scaleCmd, outFile)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error running ffmpeg (scale): %v\n%s\n", err, stderr.String())
	}
	fmt.Println("Success Scale!")
}

func calcNewSize(in imageT) imageT {
	var out imageT
	out.height = 256
	scaleValue := float32(out.height) / float32(in.height)
	out.width = int(float32(in.width) * scaleValue)
	if out.width%2 != 0 {
		out.width++
	}
	for out.width%16 != 0 {
		out.width += 2
	}
	return out
}

func cleanString(s string) imageT {
	var img imageT
	s = strings.Replace(s, "width=", "", -1)
	s = strings.Replace(s, "height=", "", -1)
	result := strings.Split(s, "\n")
	w, err := strconv.Atoi(result[0])
	if err != nil {
		log.Fatalf("Error converting string line 1: %v\n", err)
	}
	h, err := strconv.Atoi(result[1])
	if err != nil {
		log.Fatalf("Error converting string line 2: %v\n", err)
	}
	img.width = w
	img.height = h
	return img
}
