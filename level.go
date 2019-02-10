package main

import (
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/texture"
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
func NewFarm(tf *TheFarm, cam *camera.Perspective) *Stage {
	stg := new(Stage)
	stg.farm = tf
	stg.camera = cam

	stg.scene = core.NewNode()
	stg.scene.SetPosition(0, 0, 0)
	// Make Plane
	groundMaterial := material.NewPhong(math32.NewColor("Brown"))
	groundMaterial.AddTexture(NewTexture(tf.dataDir + "/assets/ground.png"))
	planeGeom := geometry.NewPlane(5, 1, 1, 1)
	makePlaneWithMaterial := func(mat *material.Phong) *graphic.Mesh {
		return graphic.NewMesh(planeGeom, mat)
	}
	mesh := makePlaneWithMaterial(groundMaterial)
	stg.scene.Add(mesh)
	log.Debug("Added Plane Mesh!")
	// Add light above the stage
	light := light.NewPoint(&math32.Color{1, 1, 1}, 8.0)
	light.SetPosition(2, 1, 2)
	stg.scene.Add(light)

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
			still_animating := anim.Update(timeDelta)
			if still_animating {
				// copy to new slice
				stg.toAnimate = append(stg.toAnimate, anim)
			}
		}
	}

}
