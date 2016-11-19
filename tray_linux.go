package systray

func _NewSystray(iconPath string, clientPath string) *_Systray {
	return &_Systray{}
}

type _Systray struct {
	lclick func()
	rclick func()
	dclick func()
}

func (p *_Systray) Show(file string, hint string) error {
	return nil
}

func (p *_Systray) Stop() error {
	return nil
}

func (p *_Systray) SetIcon(file string) error {
	return nil
}

func (p *_Systray) SetTooltip(tooltip string) error {
	return nil
}

func (p *_Systray) SetVisible(visible bool) error {
	return nil
}

func (p *_Systray) Run() error {
	return nil
}

func (p *_Systray) OnClick(fun func()) {
    p.lclick = fun
    p.rclick = fun
    p.dclick = fun
}


func (p *_Systray) ClearSystrayMenuItems() {
}

func (p *_Systray) AddSystrayMenuItems(items []CallbackInfo) {
}
