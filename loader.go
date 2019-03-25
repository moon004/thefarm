package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"time"

	"github.com/cutter"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ki"
	"github.com/pkg/errors"
	"github.com/r3s/gombine"
	"gocv.io/x/gocv"
)

var modelSelector string

// ProcessedAndSave will gombine and save
// them in tf.faceDir
func (tf *TheFarm) ProcessedAndSave() {
	fileFace, err := tf.AICam(0,
		"assets/data/ssd_model/deploy.prototxt",
		"assets/data/ssd_model/res10_300x300_ssd_iter_140000_fp16.caffemodel")
	Errs("Error processing AICam", err)

	// Here MODEL SELECTION and GOMBINE will occur.
	var fmodel string

	switch modelSelector {
	case "Son":
		fmodel = filepath.Join(tf.charDir, "Son.png")
	case "Father":
		fmodel = filepath.Join(tf.charDir, "Father.png")
	case "Mother":
		fmodel = filepath.Join(tf.charDir, "Mother.png")
	case "Daughter":
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
	Errs("Error opening file for gombine", err)
	defer fimg.Close()

	img, err := jpeg.Decode(fimg)
	Errs("Error decoding jpeg while in ImageDataGetter", err)

	imd, err := gombine.GetImageData(&img, file)
	Errs("Error getting Image Data", err)

	return imd
}

// AICam is the boilerplate for ssd-facedetection and also returns
// the cropped image.
func (tf *TheFarm) AICam(deviceID int, proto, model string) (string, error) {
	// start The Farm Gui
	go gimain.Main(func() {
		TheFarmGui()
	})

	// open capture device
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		return "",
			errors.Wrapf(err, "Error opening video capture device: %v\n",
				deviceID)
	}

	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	// open DNN classifier
	net := gocv.ReadNetFromCaffe(proto, model)
	if net.Empty() {
		fmt.Printf("Error reading network model from : %v %v\n",
			proto, model)
	}
	defer net.Close()

	window := gocv.NewWindow("The Farm")

	green := color.RGBA{0, 255, 0, 0}
	fmt.Printf("Start reading device: %v\n", deviceID)

	// FarmGui()

	//--------------------------------------//
	// The Camera For Looooooooooooop
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

		window.WaitKey(1)
		// Snap and crop here
		if modelSelector != "" {
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
			modelSelector = ""
			return saveFile, nil
		}
		window.IMShow(img)
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

func TheFarmGui() {
	width := 1024
	height := 768

	rec := ki.Node{}          // receiver for events
	rec.InitName(&rec, "rec") // this is essential for root objects not owned by other Ki tree nodes

	gi.SetAppName("widgets")
	win := gi.NewWindow2D(
		"The Farm Family",
		"The Farm Family",
		width,
		height,
		true,
	)
	vp := win.WinViewport2D()
	// style sheet
	// butwidth, butheight := "10em", "3em"
	butBGColor := gi.Color{232, 232, 232, 255}
	var css = ki.Props{
		"button": ki.Props{
			"background-color": butBGColor,
		},
		"#combo": ki.Props{
			"background-color": gi.Color{194, 232, 252, 255},
		},
		".hslides": ki.Props{
			"background-color": gi.Color{240, 225, 255, 255},
		},
		"kbd": ki.Props{
			"color": "blue",
		},
	}
	vp.CSS = css
	updt := vp.UpdateStart()
	mfr := win.SetMainFrame()
	mfr.SetProp("spacing", units.NewValue(1, units.Ex))

	// Setting row
	titlerow := gi.AddNewLayout(mfr, "titlerow", gi.LayoutHoriz)
	titlerow.SetProp("horizontal-align", gi.AlignCenter)

	buttonrow1 := gi.AddNewLayout(mfr, "buttonrow1", gi.LayoutHoriz)
	buttonrow1.SetProp("spacing", units.NewValue(3, units.Em))
	buttonrow1.SetProp("horizontal-align", gi.AlignCenter)

	descriptionRow := gi.AddNewLayout(mfr, "descRow", gi.LayoutHoriz)
	descriptionRow.SetProp("spacing", units.NewValue(3, units.Em))
	descriptionRow.SetProp("horizontal-align", gi.AlignCenter)

	snapButRow := gi.AddNewLayout(mfr, "snapButRow", gi.LayoutHoriz)
	snapButRow.SetProp("horizontal-align", gi.AlignCenter)
	snapButRow.SetProp("spacing", units.NewValue(2, units.Em))
	snapButRow.SetProp("margin", units.NewValue(2, units.Em))
	// ------------------ Title ------------------//
	title := gi.AddNewLabel(titlerow, "title", "The Farm Family")
	title.SetProp("font-size", units.NewValue(3, units.Em))
	title.SetStretchMaxWidth()
	title.SetStretchMaxHeight()

	// ------------------Descriptions-----------------//
	descSize := units.NewValue(40, units.Px)
	desc1 := gi.AddNewLabel(descriptionRow, "desc1", " Father ")
	desc1.SetProp("text-align", "center")
	desc1.SetProp("font-size", descSize)

	desc2 := gi.AddNewLabel(descriptionRow, "desc2", "     Son ")
	desc2.SetProp("text-align", "center")
	desc2.SetProp("font-size", descSize)

	desc3 := gi.AddNewLabel(descriptionRow, "desc3", "      Mother")
	desc3.SetProp("text-align", "center")
	desc3.SetProp("font-size", descSize)

	desc4 := gi.AddNewLabel(descriptionRow, "desc4", "  Daughter")
	desc4.SetProp("text-align", "center")
	desc4.SetProp("font-size", descSize)

	// ----------------- Buttons ----------------//
	iconSize := units.NewValue(10, units.Em)
	curFocus := 1

	// SnapShot Button
	butSnap := gi.AddNewButton(snapButRow, "butSnap")
	butSnap.SetIcon("camera")
	butSnap.SetProp("#icon", ki.Props{
		"width":  units.NewValue(10, units.Em),
		"height": units.NewValue(9, units.Em),
	})
	descSnap := gi.AddNewLabel(snapButRow, "descSnap", "Take Picture")
	descSnap.SetProp("font-size", descSize)
	descSnap.SetProp("vertical-align", gi.AlignCenter)

	// Family icon buttons
	button1 := gi.AddNewButton(buttonrow1, "button1")
	button1.SetIcon("father")
	button1.SetProp("#icon", ki.Props{
		"width":  iconSize,
		"height": iconSize,
	})
	button1.SetProp(":focus", ki.Props{
		"border-color":     "black",
		"border-width":     units.NewValue(8, units.Px),
		"background-color": "linear-gradient(samelight-100, highlight-20)",
	})
	button1.Tooltip = "click to select your character"

	button2 := gi.AddNewButton(buttonrow1, "button2")
	button2.SetIcon("son")
	button2.SetProp("#icon", ki.Props{
		"width":  iconSize,
		"height": iconSize,
	})
	button2.SetProp(":focus", ki.Props{
		"border-color":     "black",
		"border-width":     units.NewValue(8, units.Px),
		"background-color": "linear-gradient(samelight-100, highlight-20)",
	})
	button2.Tooltip = "click to select your character"

	button3 := gi.AddNewButton(buttonrow1, "button3")
	button3.SetIcon("mom")
	button3.SetProp("#icon", ki.Props{
		"width":  iconSize,
		"height": iconSize,
	})
	button3.SetProp(":focus", ki.Props{
		"border-color":     "black",
		"border-width":     units.NewValue(8, units.Px),
		"background-color": "linear-gradient(samelight-100, highlight-20)",
	})
	button3.Tooltip = "click to select your character"

	button4 := gi.AddNewButton(buttonrow1, "button4")
	button4.SetIcon("daughter")
	button4.SetProp("#icon", ki.Props{
		"width":  iconSize,
		"height": iconSize,
	})
	button4.SetProp(":focus", ki.Props{
		"border-color":     "black",
		"border-width":     units.NewValue(8, units.Px),
		"background-color": "linear-gradient(samelight-100, highlight-20)",
	})
	button4.Tooltip = "click to select your character"

	// -------------------- Button Click ---------------------//
	button1.ButtonSig.Connect(rec.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(gi.ButtonReleased) {
				ButStChanger(curFocus, 1, button1)
				curFocus = 1
			}
		})
	button2.ButtonSig.Connect(rec.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(gi.ButtonReleased) {
				ButStChanger(curFocus, 2, button2)
				curFocus = 2
			}
		})
	button3.ButtonSig.Connect(rec.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(gi.ButtonReleased) {
				ButStChanger(curFocus, 3, button3)
				curFocus = 3
			}
		})
	button4.ButtonSig.Connect(rec.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(gi.ButtonReleased) {
				ButStChanger(curFocus, 4, button4)
				curFocus = 4
			}
		})
	butSnap.ButtonSig.Connect(rec.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(gi.ButtonReleased) {
				fmt.Println("SnapShot!")
			}
		})
	win.MainMenuUpdated()
	vp.UpdateEndNoSig(updt)
	win.StartEventLoop()
}

func ButStChanger(curFocus, butClicked int, but *gi.Button) {
	result := curFocus - butClicked
	if result < 0 {
		for i := 0; i > result; i-- {
			but.FocusNext()
		}
	} else {
		for i := 0; i < result; i++ {
			but.FocusPrev()
		}
	}
	log.Debug("curFocus: %v ButClicked: %v", curFocus, butClicked)
}

// // FarmGui spins up the Gui and returns the selection
// // whenever a button is clicked.
// func FarmGui() {
// 	gtk.Init(nil)

// 	builder, err := gtk.BuilderNewFromFile("gtk_data/builder.ui")
// 	Errs("error loading buidler", err)

// 	// Map the handlers to callback functions, and connect the signals
// 	// to the Builder.
// 	signals := map[string]interface{}{
// 		"on_main_window_destroy": func() {
// 			log.Debug("Main window close clicked")
// 		},
// 	}
// 	builder.ConnectSignals(signals)

// 	// Get button1
// 	obj, err := builder.GetObject("window")
// 	Errs("error getting obj from builder", err)
// 	win, ok := obj.(*gtk.Window)
// 	if !ok {
// 		log.Fatal("Type assertion error")
// 	}
// 	win.SetDefaultSize(800, 300)

// 	//------------------------------------------------------//

// 	// Get button1
// 	obj, err = builder.GetObject("button1")
// 	Errs("error getting obj from builder", err)
// 	button, ok := obj.(*gtk.Button)
// 	if !ok {
// 		log.Fatal("Type assertion error")
// 	}
// 	button.Connect("clicked", func() {
// 		fmt.Println("button clicked!")
// 		modelSelector = "Father"
// 	})
// 	button.SetHExpand(true)
// 	button.SetVExpand(true)
// 	// Get button2
// 	obj, err = builder.GetObject("button2")
// 	Errs("error getting obj from builder", err)
// 	button2, ok := obj.(*gtk.Button)
// 	if !ok {
// 		log.Fatal("Type assertion error")
// 	}
// 	button2.Connect("clicked", func() {
// 		fmt.Println("button2 clicked!")
// 		modelSelector = "Son"
// 	})
// 	button2.SetHExpand(true)
// 	button2.SetVExpand(true)
// 	// Get button3
// 	obj, err = builder.GetObject("button3")
// 	Errs("error getting obj from builder", err)
// 	button3, ok := obj.(*gtk.Button)
// 	if !ok {
// 		log.Fatal("Type assertion error")
// 	}
// 	button3.Connect("clicked", func() {
// 		fmt.Println("button3 clicked!")
// 		modelSelector = "Mother"
// 	})
// 	button3.SetHExpand(true)
// 	button3.SetVExpand(true)
// 	// Get button4
// 	obj, err = builder.GetObject("button4")
// 	Errs("error getting obj from builder", err)
// 	button4, ok := obj.(*gtk.Button)
// 	if !ok {
// 		log.Fatal("Type assertion error")
// 	}
// 	button4.Connect("clicked", func() {
// 		fmt.Println("button4 clicked!")
// 		modelSelector = "Daughter"
// 	})
// 	button4.SetHExpand(true)
// 	button4.SetVExpand(true)

// 	// --------------------------------------------//

// 	// Get Grid2
// 	obj, err = builder.GetObject("image")
// 	Errs("error getting obj from buidler", err)
// 	img, ok := obj.(*gtk.Image)
// 	if !ok {
// 		log.Fatal("Type assertion error")
// 	}
// 	img.SetHExpand(true)
// 	img.SetVExpand(true)

// 	// Get title
// 	obj, err = builder.GetObject("title")
// 	Errs("error getting obj from buidler", err)
// 	title, ok := obj.(*gtk.Label)
// 	if !ok {
// 		log.Fatal("Type assertion error")
// 	}
// 	title.SetName("title")
// 	title.SetHExpand(true)
// 	title.SetVExpand(true)

// 	// Get Instruction
// 	obj, err = builder.GetObject("instruction")
// 	Errs("error getting obj from buidler", err)
// 	Instruction, ok := obj.(*gtk.Label)
// 	if !ok {
// 		log.Fatal("Type assertion error")
// 	}
// 	Instruction.SetName("Instruction")
// 	Instruction.SetHExpand(true)
// 	Instruction.SetVExpand(true)

// 	defaultScreen, err := gdk.ScreenGetDefault()
// 	Errs("error getting screen default", err)

// 	cssProvider, err := gtk.CssProviderNew()
// 	Errs("Unable to create css prov", err)

// 	err = cssProvider.LoadFromPath("gtk_data/theme.css")
// 	Errs("error loading css", err)
// 	btn, err := gtk.ButtonNewWithLabel("Button with label")
// 	Errs("Unable to create button", err)

// 	btn.SetName("button")

// 	gtk.AddProviderForScreen(defaultScreen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

// 	go gtk.Main()

// }

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
