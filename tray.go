package systray

import "runtime"

type Systray struct {
	*_Systray
}

func New(iconPath string, clientPath string) *Systray {
	st := &Systray{_NewSystray(iconPath, clientPath)}

	// Track the cleanup of the public Systray instance,
	// so that we can decref the wrapped private instance
	runtime.SetFinalizer(st, func(ptr *Systray) {
		if ptr._Systray != nil {
			gSystrays.Decref(ptr._Systray.refId)
		}
	})

	return st
}

type CallbackInfo struct {
	ItemName string
	Callback func()
	Disabled bool
	Checked  bool
}
