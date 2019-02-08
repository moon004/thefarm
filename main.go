package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/g3n/engine/audio"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/camera/control"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/logger"
	"github.com/g3n/engine/window"
)

var log *logger.Logger

type TheFarm struct {
	wmgr     window.IWindowManager
	win      window.IWindow
	gs       *gls.GLS
	renderer *renderer.Renderer
	scene    *core.Node
	camera   *camera.Perspective
	dataDir  string

	userData  *UserData
	stepDelta *math32.Vector2

	levelScene     *core.Node
	charNode       *core.Node
	audioAvailable bool

	//Sound and Sfx
	musicPlayer   *audio.Player
	charCreateSnd *audio.Player
}

func (farm *TheFarm) ResetFarm(playsound bool) {
	log.Debug("Reset Farm")

	farm.charNode = 0

}

// ToggleFullScreen toggles whether is game is fullscreen or windowed
func (g *TheFarm) ToggleFullScreen() {
	log.Debug("Toggle FullScreen")

	g.win.SetFullScreen(!g.win.FullScreen())
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
	f := new(TheFarm)

	// Manually scan the $GOPATH directories to find the data directory
	rawPaths := os.Getenv("GOPATH")
	paths := strings.Split(rawPaths, ":")
	for _, j := range paths {
		// Checks data path
		path := filepath.Join(j, "src", "github.com", "moon004", "TheFarm")
		if _, err := os.Stat(path); err == nil {
			f.dataDir = path
		}
		Errs(err)
	}

	// Load user data from file
	f.userData = NewUserData(f.dataDir)

	// Get the window manager
	var err error
	f.wmgr, err = window.Manager("glfw")
	if err != nil {
		panic(err)
	}

	// Create window and OpenGL context
	f.win, err = f.wmgr.CreateWindow(1200, 900, "Farm", f.userData.FullScreen)
	if err != nil {
		panic(err)
	}

	// Create OpenGL state
	f.gs, err = gls.New()
	if err != nil {
		panic(err)
	}

	// Speed up a bit by not checking OpenGL errors
	f.gs.SetCheckErrors(false)

	// Sets window background color
	f.gs.ClearColor(0.1, 0.1, 0.1, 1.0)

	// Sets the OpenGL viewport size the same as the window size
	// This normally should be updated if the window is resized.
	width, height := f.win.Size()
	f.gs.Viewport(0, 0, int32(width), int32(height))

	// Creates GUI root panel
	f.root = gui.NewRoot(f.gs, f.win)
	f.root.SetSize(float32(width), float32(height))

	// Subscribe to window resize events. When the window is resized:
	// - Update the viewport size
	// - Update the root panel size
	// - Update the camera aspect ratio
	f.win.Subscribe(window.OnWindowSize, func(evname string, ev interface{}) {
		width, height := f.win.Size()
		f.gs.Viewport(0, 0, int32(width), int32(height))
		f.root.SetSize(float32(width), float32(height))
		aspect := float32(width) / float32(height)
		f.camera.SetAspect(aspect)
	})

	// Subscribe window to events
	f.win.Subscribe(window.OnKeyDown, f.onKey)
	f.win.Subscribe(window.OnMouseUp, f.onMouse)
	f.win.Subscribe(window.OnMouseDown, f.onMouse)

	// Creates a renderer and adds default shaders
	f.renderer = renderer.NewRenderer(f.gs)
	//f.renderer.SetSortObjects(false)
	err = f.renderer.AddDefaultShaders()
	if err != nil {
		panic(err)
	}
	f.renderer.SetGui(f.root)

	// Adds a perspective camera to the scene
	// The camera aspect ratio should be updated if the window is resized.
	aspect := float32(width) / float32(height)
	f.camera = camera.NewPerspective(65, aspect, 0.01, 1000)
	f.camera.SetPosition(0, 4, 5)
	f.camera.LookAt(&math32.Vector3{0, 0, 0})

	// Create orbit control and set limits
	f.orbitControl = control.NewOrbitControl(f.camera, f.win)
	f.orbitControl.Enabled = false
	f.orbitControl.EnablePan = false
	f.orbitControl.MaxPolarAngle = 2 * math32.Pi / 3
	f.orbitControl.MinDistance = 5
	f.orbitControl.MaxDistance = 15

	// Create main scene and child levelScene
	f.scene = core.NewNode()
	f.levelScene = core.NewNode()
	f.scene.Add(f.camera)
	f.scene.Add(f.levelScene)
	f.stepDelta = math32.NewVector2(0, 0)
	f.renderer.SetScene(f.scene)

	// Add white ambient light to the scene
	ambLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.4)
	f.scene.Add(ambLight)

	f.levelStyle = NewStandardStyle(f.dataDir)

	f.SetupGui(width, height)
	f.RenderFrame()

	// Try to open audio libraries
	err = loadAudioLibs()
	if err != nil {
		lof.Error("%s", err)
		f.UpdateMusicButton(false)
		f.UpdateSfxButton(false)
		f.musicButton.SetEnabled(false)
		f.sfxButton.SetEnabled(false)
	} else {
		f.audioAvailable = true
		f.LoadAudio()
		f.UpdateMusicButton(f.userData.MusicOn)
		f.UpdateSfxButton(f.userData.SfxOn)

		// Queue the music!
		f.musicPlayerMenu.Play()
	}

	f.LoadSkyBox()
	f.LoadGopher()
	f.CreateArrowNode()
	f.LoadLevels()

	f.win.Subscribe(window.OnCursor, f.onCursor)

	if f.userData.LastUnlockedLevel == len(f.levels) {
		f.titleImage.SetImage(gui.ButtonDisabled, f.dataDir+"/gui/title3_completed.png")
	}

	// Done Loading - hide the loading label, show the menu, and initialize the level
	f.loadingLabel.SetVisible(false)
	f.menu.Add(f.main)
	f.InitLevel(f.userData.LastLevel)
	f.gopherLocked = true

	now := time.Now()
	newNow := time.Now()
	lof.Info("Starting Render Loop")

	// Start the render loop
	for !f.win.ShouldClose() {

		newNow = time.Now()
		timeDelta := now.Sub(newNow)
		now = newNow

		f.Update(timeDelta.Seconds())
		f.RenderFrame()
	}
}
