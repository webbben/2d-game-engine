package popup

import "github.com/hajimehoshi/ebiten/v2"

type Manager struct {
	currentPopup Popupable
}

func NewPopupManager() Manager {
	m := Manager{}
	return m
}

func (m Manager) IsPopupActive() bool {
	return m.currentPopup != nil
}

type Popupable interface {
	Draw(screen *ebiten.Image)
	Update()
	IsClosed() bool // this should be true once the popup has closed and we are done with it.
}

func (m *Manager) SetPopup(newPopup Popupable) {
	if newPopup == nil {
		panic("popup is nil")
	}
	if m.currentPopup != nil {
		panic("popup is already set. make sure a previous popup has been cleared before setting a new one.")
	}
	m.currentPopup = newPopup
}

func (m *Manager) ClearPopup() {
	m.currentPopup = nil
}

func (m *Manager) Update() {
	if m.currentPopup == nil {
		return
	}
	if m.currentPopup.IsClosed() {
		m.ClearPopup()
		return
	}
	m.currentPopup.Update()
}

func (m *Manager) Draw(screen *ebiten.Image) {
	if m.currentPopup == nil {
		return
	}
	m.currentPopup.Draw(screen)
}
