package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/g3n/engine/audio/al"
	"github.com/g3n/engine/audio/vorbis"

	"github.com/g3n/engine/audio"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/camera/control"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/loader/obj"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/logger"
	"github.com/g3n/engine/window"
	"github.com/pkg/errors"
)

var log *logger.Logger

// TheFarm is the main struct of the application
type TheFarm struct {
	wmgr         window.IWindowManager
	win          window.IWindow
	gs           *gls.GLS
	renderer     *renderer.Renderer
	scene        *core.Node
	camera       *camera.Perspective
	orbitControl *control.OrbitControl
	dataDir      string

	userData  *UserData
	stepDelta *math32.Vector2

	stageScene     *core.Node
	stage          *Stage
	charNode       *core.Node
	audioAvailable bool

	//Sound and Sfx
	musicPlayer   *audio.Player
	charCreateSnd *audio.Player
}

// ResetFarm clears all the characters.
func (tf *TheFarm) ResetFarm() {
	log.Debug("Reset Farm")

	tf.charNode = nil

}

// ToggleFullScreen toggles whether is game is fullscreen or windowed
func (tf *TheFarm) ToggleFullScreen() {
	log.Debug("Toggle FullScreen")

	tf.win.SetFullScreen(!tf.win.FullScreen())
}

// Update updates the current stage if any
func (tf *TheFarm) Update(timeDelta float64) {
	if tf.stage != nil {
		tf.stage.Update(timeDelta)
	}
}

// onKey handles key R and key Enter
func (tf *TheFarm) onKey(evname string, ev interface{}) {
	kev := ev.(*window.KeyEvent) // return key events
	switch kev.Keycode {
	case window.KeyR:
		tf.ToggleFullScreen()
	case window.KeyEnter:
		tf.ResetFarm()
	}
}

func (tf *TheFarm) onCursor(evname string, ev interface{}) {
	var dir math32.Vector3

	tf.camera.WorldDirection(&dir)
	tf.stepDelta.Set(0, 0)

	if math32.Abs(dir.Z) > math32.Abs(dir.X) {
		if dir.Z > 0 {
			tf.stepDelta.Y = 1
		} else {
			tf.stepDelta.Y = -1
		}
	} else {
		if dir.X > 0 {
			tf.stepDelta.X = 1
		} else {
			tf.stepDelta.X = -1
		}
	}
}

// CreateChar creates character and add it to the Scene
// in g3n, Scene is actually *core.Node and adding
// *core.Node is actually adding object to Scene
func (tf *TheFarm) CreateChar(txName, name string) {
	log.Debug("Creating Character")
	//(rad, widthseg, heighseg, phist, philen, thetast, thetalen)
	// geom := geometry.NewSphere(1, 10, 10, 0, 3, 0, 3)

	// // adding the texture to the shape
	// mat := material.NewPhong(math32.NewColor("White"))
	// mat.AddTexture(NewTexture(txName))

	// sphere := graphic.NewMesh(geom, mat)
	// sphere.SetName(name)
	// sphere.SetPosition(0, 0, 0)
	// sphere.SetRotation(0, 0, 3.14159)

	// tf.charNode = core.NewNode()
	// tf.charNode.Add(sphere)
	// tf.stageScene.Add(tf.charNode)

	dec, err := obj.Decode(tf.dataDir+"/face/char.obj", tf.dataDir+"/face/char.mtl")
	Errs(err)

	char, err := dec.NewGroup()
	Errs(err)

	tf.charNode = core.NewNode()
	tf.charNode.Add(char)
	tf.stageScene.Add(tf.charNode)
	log.Debug("New character CREATED!")

}

// LoadStage loads the stage and put inside tf.stage
func (tf *TheFarm) LoadStage() {
	log.Debug("Loading Stage")

	// TODO load stage model from Blender

	tf.stage = NewFarm(tf, tf.camera)
	tf.stageScene.Add(tf.stage.scene)
	// allow camera movement
	tf.orbitControl.Enabled = true
}

// loadAudioLibs
func loadAudioLibs() error {

	// Open default audio device
	dev, err := al.OpenDevice("")
	Errs(errors.Wrap(err, "Error opening OpenAL default device"))

	// Create audio context
	acx, err := al.CreateContext(dev, nil)
	Errs(errors.Wrap(err, "Error creating audio context"))

	// Make the context the current one
	err = al.MakeContextCurrent(acx)
	Errs(errors.Wrap(err, "Error setting audio context current"))
	log.Debug("%s version: %s", al.GetString(al.Vendor), al.GetString(al.Version))
	log.Debug("%s", vorbis.VersionString())
	return nil
}

// LoadAudio loads music and sound effects
func (tf *TheFarm) LoadAudio() {
	log.Debug("Load Audio")

	// Create listener and add it to the current camera
	listener := audio.NewListener()
	cdir := tf.camera.Direction()
	listener.SetDirectionVec(&cdir)
	tf.camera.GetCamera().Add(listener)

	// Helper function to create player and handle errors
	createPlayer := func(fname string) *audio.Player {
		log.Debug("Loading " + fname)
		p, err := audio.NewPlayer(fname)
		Errs(errors.Wrapf(err, "Failed to create player for: %v", fname))
		return p
	}

	tf.musicPlayer = createPlayer(tf.dataDir + "/BGM.ogg")
	tf.musicPlayer.SetLooping(true)
}

// PlaySound just play the sound by:
// PlaySound(tf.musicPlayer, nil)
func (tf *TheFarm) PlaySound(player *audio.Player, node *core.Node) {
	if tf.audioAvailable {
		if node != nil {
			node.Add(player)
		}
		player.Stop()
		player.Play()
	}
}

func main() {
	// OpenGL functions must be executed in the same thread where
	// the context was created (by window.New())
	runtime.LockOSThread()

	// Parse command line flags
	showLog := flag.Bool("debug", false, "display the debug log")
	flag.Parse()

	// Create logger
	log = logger.New("TheFarm", nil)
	log.AddWriter(logger.NewConsole(false))
	log.SetFormat(logger.FTIME | logger.FMICROS)
	if *showLog == true {
		log.SetLevel(logger.DEBUG)
	}

	// Create TheFarm struct
	tf := new(TheFarm)

	// Manually scan the $GOPATH directories to find the data directory
	rawPaths := os.Getenv("GOPATH")
	paths := strings.Split(rawPaths, ":")
	for _, j := range paths {
		// Checks data path
		path := filepath.Join(j, "src", "github.com", "louis-project", "assets")
		if _, err := os.Stat(path); err == nil {
			tf.dataDir = path
		}
	}

	// Load user data from file
	// userData {
	// MusicOn    bool
	// SfxOn      bool
	// MusicVol   float32
	// SfxVol     float32
	// FullScreen bool
	// }
	tf.userData = NewUserData(tf.dataDir)

	// Get the window manager
	var err error
	tf.wmgr, err = window.Manager("glfw")
	Errs(err)

	// Create window and OpenGL context
	tf.win, err = tf.wmgr.CreateWindow(1200, 900, "Farm", tf.userData.FullScreen)
	Errs(err)

	// Create OpenGL state
	tf.gs, err = gls.New()
	Errs(err)

	// Speed up a bit by not checking OpenGL errors
	tf.gs.SetCheckErrors(false)

	// Sets window background color
	tf.gs.ClearColor(0.1, 0.1, 0.1, 1.0)

	// Sets the OpenGL viewport size the same as the window size
	// This normally should be updated if the window is resized.
	width, height := tf.win.Size()
	tf.gs.Viewport(0, 0, int32(width), int32(height))

	// Subscribe to window resize events. When the window is resized:
	// - Update the viewport size
	// - Update the root panel size
	// - Update the camera aspect ratio
	tf.win.Subscribe(window.OnWindowSize, func(evname string, ev interface{}) {
		width, height := tf.win.Size()
		tf.gs.Viewport(0, 0, int32(width), int32(height))
		aspect := float32(width) / float32(height)
		tf.camera.SetAspect(aspect)
	})

	// Subscribe window to events
	tf.win.Subscribe(window.OnKeyDown, tf.onKey)

	// Creates a renderer and adds default shaders
	tf.renderer = renderer.NewRenderer(tf.gs)
	//tf.renderer.SetSortObjects(false)
	err = tf.renderer.AddDefaultShaders()
	Errs(err)

	// Adds a perspective camera to the scene
	// The camera aspect ratio should be updated if the window is resized.
	aspect := float32(width) / float32(height)
	tf.camera = camera.NewPerspective(65, aspect, 0.01, 1000)
	tf.camera.SetPosition(0, 4, 5)
	tf.camera.LookAt(&math32.Vector3{0, 0, 0})

	// Create orbit control and set limits
	tf.orbitControl = control.NewOrbitControl(tf.camera, tf.win)
	tf.orbitControl.Enabled = false
	tf.orbitControl.EnablePan = false
	tf.orbitControl.MaxPolarAngle = 2 * math32.Pi / 3
	tf.orbitControl.MinDistance = 5
	tf.orbitControl.MaxDistance = 15

	// Create main scene and child stageScene
	tf.scene = core.NewNode()
	tf.stageScene = core.NewNode()
	tf.scene.Add(tf.camera)
	tf.scene.Add(tf.stageScene)
	tf.stepDelta = math32.NewVector2(0, 0)
	tf.renderer.SetScene(tf.scene)

	// Add white ambient light to the scene
	ambLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.4)
	tf.scene.Add(ambLight)

	tf.RenderFrame()

	// Try to open audio libraries
	err = loadAudioLibs()
	if err != nil {
		Errs(err)
	} else {
		tf.audioAvailable = true
		tf.LoadAudio()
		tf.musicPlayer.Play()
	}
	tf.CreateChar(tf.dataDir+"/face/f1.png", "sphere")
	tf.LoadStage()

	tf.win.Subscribe(window.OnCursor, tf.onCursor)

	now := time.Now()
	newNow := time.Now()
	log.Debug("Starting Render Loop")

	// Start the render loop
	for !tf.win.ShouldClose() {
		newNow = time.Now()
		timeDelta := now.Sub(newNow)
		now = newNow

		tf.Update(timeDelta.Seconds())
		tf.RenderFrame()
	}
}

// RenderFrame renders a frame of the scene with the GUI overlaid
func (tf *TheFarm) RenderFrame() {

	// Render the scene/gui using the specified camera
	rendered, err := tf.renderer.Render(tf.camera)
	Errs(err)

	// Check I/O events
	tf.wmgr.PollEvents()

	// Update window if necessary
	if rendered {
		tf.win.SwapBuffers()
	}
}
