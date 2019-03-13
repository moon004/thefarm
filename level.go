package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/texture"
	"github.com/pkg/errors"
)

// Stage struct
type Stage struct {
	farm   *TheFarm
	scene  *core.Node
	camera *camera.Perspective

	toAnimate []*Animation
	animating bool
	resetAnim bool
}

// NewFarm Creates new stage for it
// func NewFarm(tf *TheFarm, cam *camera.Perspective) *Stage {
// 	stg := new(Stage)
// 	stg.farm = tf
// 	stg.camera = cam

// 	stg.scene = core.NewNode()
// 	stg.scene.SetPosition(0, 0, 0)
// 	// Make Plane
// 	groundMaterial := material.NewPhong(math32.NewColor("Brown"))
// 	groundMaterial.AddTexture(NewTexture(tf.dataDir + "/ground.png"))
// 	planeGeom := geometry.NewPlane(5, 1, 5, 5)
// 	mesh := graphic.NewMesh(planeGeom, groundMaterial)
// 	mesh.SetRotation(-1.5708, 0, 0)
// 	stg.scene.Add(mesh)
// 	log.Debug("Added Plane Mesh!")

// 	return stg
// }

// NewStage returns a loaded *Stage (pointer to stage)
func NewStage(tf *TheFarm, cam *camera.Perspective) *Stage {
	stg := new(Stage)
	stg.farm = tf
	stg.camera = cam

	stg.scene = core.NewNode()
	stg.scene.SetPosition(0, 0, 0)
	// Make Plane

	// Load Stage no need faceID
	files, err := ioutil.ReadDir(tf.stageDir)
	Errs(errors.WithStack(err))

	for _, f := range files {
		//Get File extension, if gltf, add it to scene
		ext := filepath.Ext(f.Name())

		if ext == ".gltf" {
			file := filepath.Join(tf.stageDir, f.Name())
			node := tf.loadScene(file, "")
			stg.scene.Add(node)
		}
	}
	// node := tf.loadScene(tf.stageDir, "")
	// stg.scene.Add(node)
	log.Debug("Added Stage Farm!")

	return stg
}

// NewTexture returns new *texture.Texture2D
func NewTexture(path string) *texture.Texture2D {
	tex, err := texture.NewTexture2DFromImage(path)
	if err != nil {
		log.Fatal("Error loading texture: %s", err)
	}
	return tex
}

// Update updates all ongoing animations for the level
func (stg *Stage) Update(timeDelta float64) {

	if stg.resetAnim {
		stg.resetAnim = false
		stg.toAnimate = make([]*Animation, 0)
	}

	newToAnimate := stg.toAnimate
	stg.toAnimate = make([]*Animation, 0)

	for _, anim := range newToAnimate {
		if !stg.resetAnim {
			stillAnimating := anim.Update(timeDelta)
			if stillAnimating {
				// copy to new slice
				stg.toAnimate = append(stg.toAnimate, anim)
			}
		}
	}

}
