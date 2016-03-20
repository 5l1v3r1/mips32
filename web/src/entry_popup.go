package main

import (
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

type entryPopup struct {
	shieldElement *js.Object
	popupElement  *js.Object
	inputElement  *js.Object
	callback      func(i uint32)
}

func NewEntryPopup(prompt string, callback func(i uint32)) {
	e := &entryPopup{}

	document := js.Global.Get("document")

	e.shieldElement = document.Call("createElement", "div")
	e.shieldElement.Set("className", "popup-shield")

	e.popupElement = document.Call("createElement", "div")
	e.popupElement.Set("className", "popup")

	titleElement := document.Call("createElement", "label")
	titleElement.Set("className", "popup-prompt")
	titleElement.Set("innerText", prompt)
	e.popupElement.Call("appendChild", titleElement)

	e.inputElement = document.Call("createElement", "input")
	e.inputElement.Set("className", "popup-input")
	e.popupElement.Call("appendChild", e.inputElement)

	br := document.Call("createElement", "br")
	e.popupElement.Call("appendChild", br)

	cancelButton := document.Call("createElement", "button")
	cancelButton.Set("className", "popup-button popup-cancel")
	cancelButton.Set("innerText", "Cancel")
	e.popupElement.Call("appendChild", cancelButton)

	okButton := document.Call("createElement", "button")
	okButton.Set("className", "popup-button popup-ok")
	okButton.Set("innerText", "OK")
	e.popupElement.Call("appendChild", okButton)

	e.shieldElement.Call("addEventListener", "click", e.close)
	cancelButton.Call("addEventListener", "click", e.close)
	okButton.Call("addEventListener", "click", e.ok)

	document.Get("body").Call("appendChild", e.shieldElement)
	document.Get("body").Call("appendChild", e.popupElement)

	size := e.popupElement.Get("offsetHeight").Float()
	halfSize := int(size / 2)
	e.popupElement.Get("style").Set("top", "calc(50% - "+strconv.Itoa(halfSize)+"px)")

	e.callback = callback
}

func (e *entryPopup) close() {
	document := js.Global.Get("document")
	document.Get("body").Call("removeChild", e.shieldElement)
	document.Get("body").Call("removeChild", e.popupElement)
}

func (e *entryPopup) ok() {
	value := e.inputElement.Get("value").String()
	number, _ := strconv.ParseInt(value, 0, 64)
	e.close()
	e.callback(uint32(number))
}
