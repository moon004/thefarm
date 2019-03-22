package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"time"

	"github.com/r3s/gombine"

	"github.com/cutter"
	"github.com/pkg/errors"
	"gocv.io/x/gocv"
)

// ProcessedAndSave will gombine and save
// them in tf.faceDir
func (tf *TheFarm) ProcessedAndSave() {
	fileFace, err := tf.AICam(0, "assets/data/deploy.prototxt",
		"assets/data/haarcascade_frontalface_default.xml")
	Errs(err)

	// Here MODEL SELECTION and GOMBINE will occur.
	var fmodel string
	selection := "Son"
	switch selection {
	case "Son":
		fmodel = filepath.Join(tf.charDir, "Son.png")
	case "Father":
		fmodel = filepath.Join(tf.charDir, "Father.png")
	case "Mother":
		fmodel = filepath.Join(tf.charDir, "Mother.png")
	case "Daughther":
		fmodel = filepath.Join(tf.charDir, "Mother.png")
	}

	images := []*gombine.ImageData{}
	imdModel := ImageDataGetter(fmodel)
	imdFace := ImageDataGetter(fileFace)
	images = append(images, &imdModel, &imdFace)
	gombine.ProcessImages(images, "jpg", "bottom", fileFace)
}

// ImageDataGetter returns image data from the provided file
func ImageDataGetter(file string) gombine.ImageData {
	fimg, err := os.Open(file)
	Errs(err)
	defer fimg.Close()

	img, err := jpeg.Decode(fimg)
	Errs(err)

	imd, err := gombine.GetImageData(&img, file)
	Errs(err)

	return imd
}

// AICam is the boilerplate for ssd-facedetection and also returns
// the cropped image.
func (tf *TheFarm) AICam(deviceID int, proto, model string) (string, error) {

	// open capture device
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		return "",
			errors.Wrapf(err, "Error opening video capture device: %v\n",
				deviceID)
	}

	defer webcam.Close()

	window := gocv.NewWindow("SSD Face Detection")
	defer window.Close()

	img := gocv.NewMat()
	defer img.Close()

	// open DNN classifier
	net := gocv.ReadNetFromCaffe(proto, model)
	if net.Empty() {
		fmt.Printf("Error reading network model from : %v %v\n",
			proto, model)
	}
	defer net.Close()

	green := color.RGBA{0, 255, 0, 0}
	fmt.Printf("Start reading device: %v\n", deviceID)

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return "", nil
		}
		if img.Empty() {
			continue
		}

		W := float32(img.Cols())
		H := float32(img.Rows())

		// convert image Mat to 96x128 blob that the detector can analyze
		blob := gocv.BlobFromImage(img,
			1.0,
			image.Pt(128, 96),
			gocv.NewScalar(104.0, 177.0, 123.0, 0),
			false,
			false,
		)
		defer blob.Close()

		// feed the blob into the classifier
		net.SetInput(blob, "data")

		// run a forward pass through the network
		detBlob := net.Forward("detection_out")
		defer detBlob.Close()

		// extract the detections.
		// for each object detected, there will be 7 float features:
		// objid, classid, confidence, left, top, right, bottom.
		detections := gocv.GetBlobChannel(detBlob, 0, 0)
		defer detections.Close()

		var rect image.Rectangle

		for r := 0; r < detections.Rows(); r++ {
			// you would want the classid for general object detection,
			// but we do not need it here.
			// classid := detections.GetFloatAt(r, 1)

			confidence := detections.GetFloatAt(r, 2)
			if confidence < 0.3 {
				continue
			}

			left := detections.GetFloatAt(r, 3) * W
			top := detections.GetFloatAt(r, 4) * H
			right := detections.GetFloatAt(r, 5) * W
			bottom := detections.GetFloatAt(r, 6) * H

			// scale to video size:
			left = min(max(0, left), W-1)
			right = min(max(0, right), W-1)
			bottom = min(max(0, bottom), H-1)
			top = min(max(0, top), H-1)

			// draw it
			rect = image.Rect(int(left), int(top), int(right), int(bottom))
			gocv.Rectangle(&img, rect, green, 3)
		}
		window.IMShow(img)

		// Snap and crop here
		if window.WaitKey(1) >= 0 {
			tmp := "tmp.jpg"
			CT := time.Now()
			gocv.IMWrite(tmp, img)
			croppedImg, err := Cropper(rect, tmp)
			if err != nil {
				return "", err
			}

			// Write out the file (image)
			saveFile := fmt.Sprintf(filepath.Join(tf.faceDir, "%v:%v:%v.jpg"),
				CT.Hour(), CT.Minute(), CT.Second())
			filewriter, err := os.Create(saveFile)
			jpeg.Encode(filewriter, croppedImg, &jpeg.Options{
				Quality: 100, // Best quality
			})
			return saveFile, nil
		}
	}
}

// Cropper crop the saved image from gocv
func Cropper(rect image.Rectangle, tmp string) (image.Image, error) {
	file, err := os.Open(tmp)
	if err != nil {
		return nil, errors.Wrap(err, "Error opening file")
	}
	if err != nil {
		return nil, errors.Wrap(err, "Error creating file")
	}

	dec, err := jpeg.Decode(file)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding jpg")
	}
	croppedImg, err := cutter.Crop(dec, cutter.Config{
		Width:  rect.Dx(),
		Height: rect.Dy(),
		Anchor: image.Point{rect.Min.X, rect.Min.Y},
		Mode:   cutter.TopLeft,
	})

	return croppedImg, nil
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
