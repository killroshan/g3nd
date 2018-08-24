package gui

import (
	"github.com/g3n/g3nd/demos"
	"github.com/g3n/g3nd/app"
	"github.com/g3n/engine/gui"

)

type btn struct {}

func init(){
	demos.Map["gui.btn"] = &btn{}
}

func (t *btn)Initialize(app * app.App){
	b := gui.NewButton("button")
	b.SetPosition(10, 10)
	app.GuiPanel().Add(b)
}

func (t *btn)Render(app *app.App){

}


