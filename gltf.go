package main

import (
	"path/filepath"
	"strings"

	"github.com/g3n/engine/core"

	"github.com/g3n/engine/animation"
	"github.com/g3n/engine/loader/gltf"
	"github.com/g3n/g3nd/util"
	"github.com/pkg/errors"
)

type GltfLoader struct {
	prevLoaded core.INode
	selFile    *util.FileSelectButton
	anims      []*animation.Animation
}

// Render is to update gltf animation
func (tf *TheFarm) Render(delta float32) {

	for _, anim := range tf.anims {
		anim.Update(delta)
	}
}

func (tf *TheFarm) loadScene(fpath, faceID string) core.INode {

	// TODO move camera or scale scene such that it's nicely framed
	// TODO do this for other loaders as well

	// Checks file extension
	ext := filepath.Ext(fpath)
	var g *gltf.GLTF
	var err error
	// Parse the fpath directory
	item := strings.Split(fpath, "/")
	// Pick second last for item load type and last for faces
	itemToLoad := item[len(item)-2]

	// Parses file
	if ext == ".gltf" {
		g, err = gltf.ParseJSON(fpath)
		// REMEMBER ADD user facial picture HERE!!!!
		// g.Images[0].Uri = "CesiumMan1.jpg"
		switch itemToLoad {
		case "face":
			g.Images[0].Uri = filepath.Join(tf.faceDir, faceID)
		default: // Other than "character"
			log.Debug("Default case means to load stage")
		}
	} else if ext == ".glb" {
		g, err = gltf.ParseBin(fpath)
	} else {
		Errs(errors.Errorf("Uknown file extension:%s", ext))
	}
	Errs(err)

	defaultSceneIdx := 0
	if g.Scene != nil {
		defaultSceneIdx = *g.Scene
	}

	// Create default scene
	n, err := g.LoadScene(defaultSceneIdx)
	Errs(err)

	// Create animations
	for i := range g.Animations {
		anim, _ := g.LoadAnimation(i)
		anim.SetLoop(true)
		tf.anims = append(tf.anims, anim)
	}

	return n

}
