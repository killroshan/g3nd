package loader

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/graphic"
	"path/filepath"

	"github.com/g3n/engine/animation"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/loader/gltf"
	"github.com/g3n/engine/math32"
	"github.com/g3n/g3nd/app"
	"github.com/g3n/g3nd/demos"
	"github.com/g3n/g3nd/util"
	"github.com/golang/glog"
)

func init() {
	demos.Map["loader.gltf"] = &GltfLoader{}
}

type GltfLoader struct {
	prevLoaded     core.INode
	selFile        *util.FileSelectButton
	anims          []*animation.Animation
	animationGroup *gui.ControlFolderGroup
	emotesGroup    *gui.ControlFolderGroup
}

func (t *GltfLoader) Initialize(a *app.App) {

	// Creates file selection button
	t.selFile = util.NewFileSelectButton(a.DirData()+"/gltf", "Select File", 400, 300)
	t.selFile.SetPosition(10, 10)
	t.selFile.FS.SetFileFilters("*.gltf", "*.glb")
	a.GuiPanel().Add(t.selFile)
	t.selFile.Subscribe("OnSelect", func(evname string, ev interface{}) {
		fpath := ev.(string)
		err := t.loadScene(a, fpath)
		if err == nil {
			t.selFile.Label.SetText("File: " + filepath.Base(fpath))
			t.selFile.SetError("")
		} else {
			t.selFile.Label.SetText("Select File")
			t.selFile.SetError(err.Error())
		}
	})

	// Adds white directional front light
	l1 := light.NewDirectional(math32.NewColor("white"), 1.0)
	l1.SetPosition(0, 0, 10)
	a.Scene().Add(l1)

	// Adds white directional top light
	l2 := light.NewDirectional(math32.NewColor("white"), 1.0)
	l2.SetPosition(0, 10, 0)
	a.Scene().Add(l2)

	// Adds white directional right light
	l3 := light.NewDirectional(math32.NewColor("white"), 1.0)
	l3.SetPosition(10, 0, 0)
	a.Scene().Add(l3)

	// Adds axis helper
	axis := graphic.NewAxisHelper(2)
	a.Scene().Add(axis)

	// Label for error message
	errLabel := gui.NewLabel("")
	errLabel.SetFontSize(18)
	a.Gui().Add(errLabel)

	t.animationGroup = a.ControlFolder().AddGroup("Animations")
	t.emotesGroup    = a.ControlFolder().AddGroup("Emotes")

	//fpath := "gltf/DamagedHelmet/glTF/DamagedHelmet.gltf"
	fpath := "gltf/RobotExpressive.glb"
	t.loadScene(a, filepath.Join(a.DirData(), fpath))
	t.selFile.Label.SetText("File: " + filepath.Base(fpath))

}

func (t *GltfLoader) Render(a *app.App) {

	for _, anim := range t.anims {
		anim.Update(a.FrameDeltaSeconds())
	}
}

func (t *GltfLoader) loadScene(a *app.App, fpath string) error {

	// TODO move camera or scale scene such that it's nicely framed
	// TODO do this for other loaders as well

	// Remove previous model from the scene
	if t.prevLoaded != nil {
		t.anims = t.anims[:0]
		a.Scene().Remove(t.prevLoaded)
		t.prevLoaded.Dispose()
		t.prevLoaded = nil
	}

	// Checks file extension
	ext := filepath.Ext(fpath)
	var g *gltf.GLTF
	var err error

	// Parses file
	if ext == ".gltf" {
		g, err = gltf.ParseJSON(fpath)
	} else if ext == ".glb" {
		g, err = gltf.ParseBin(fpath)
	} else {
		return fmt.Errorf("Unrecognized file extension:%s", ext)
	}

	if err != nil {
		return err
	}

	spew.Config.Indent = "   "
	//spew.Dump(g.Nodes)
	//spew.Dump(g.Meshes)
	//spew.Dump(g.Accessors)

	defaultSceneIdx := 0
	if g.Scene != nil {
		defaultSceneIdx = *g.Scene
	}

	// Create default scene
	n, err := g.NewScene(defaultSceneIdx)
	if err != nil {
		return err
	}

	// Create animations
	for i := range g.Animations {
		anim, err := g.NewAnimation(i)
		if err != nil {
			glog.Error(err)
			continue
		}
		anim.SetLoop(true)
		t.anims = append(t.anims, anim)
	}

	for i := range g.Skins {
		_, err := g.NewSkeleton(i)
		if err != nil {
			glog.Error(err)
		}
	}

	g.BindSkeletion()

	t.animationGroup.RemoveAll()
	for idx, anim := range t.anims {
		idx := idx
		anim.SetPaused(true)
		cb := gui.NewCheckBox(anim.Name())
		t.animationGroup.AddPanel(cb)
		cb.Subscribe(gui.OnChange, func(name string, ev interface{}) {
			paused := t.anims[idx].Paused()
			t.anims[idx].SetPaused(!paused)
		})
	}

	// Add normals helper
	//box := n.GetNode().Children()[0].GetNode().Children()[0]
	//normals := graphic.NewNormalsHelper(box.(graphic.IGraphic), 0.1, &math32.Color{0, 0, 1}, 1)
	//a.Scene().Add(normals)

	a.Scene().Add(n)

	t.emotesGroup.RemoveAll()
	mgs := make([]*geometry.MorphGeometry, 0)
	a.Scene().OperateOnChildren(func(node core.INode) {
		if node.GetNode().Name() == "Head[1/3]" {
			fmt.Println(node.GetNode().Name())
		}

		if igr, ok := node.(graphic.IGraphic); ok {
			igeo := igr.IGeometry()
			if mg, ok := igeo.(*geometry.MorphGeometry); ok {
				mgs = append(mgs, mg)
			}
		}
	})

	if len(mgs) > 0 {
		emotes_num := len(mgs[0].GetTargets())
		weights := make([]float32, emotes_num, emotes_num)

		for idx := 0; idx < emotes_num; idx++ {
			slide := a.ControlFolder().AddSlider(fmt.Sprintf("%d", idx), 1, 0)
			idx := idx
			slide.Subscribe(gui.OnChange, func(name string, ev interface{}) {
				ref := &weights[idx]
				*ref = slide.Value()
				for _, mg := range(mgs) {
					mg.SetWeights(weights)
				}
			})
		}
	}

	t.prevLoaded = n
	return nil
}
