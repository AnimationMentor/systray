package systray

import "runtime"

type Systray struct {
	// platform-specific implementation
	systrayer
}

// systrayer interface represents the public
// and private interface needed for a platform
// specific systray implementation
type systrayer interface {
	Show(file, hint string) error
	Stop() error
	SetIcon(file string) error
	SetTooltip(tooltip string) error
	SetVisible(visible bool) error
	Run() error
	OnClick(fun func())
	ClearSystrayMenuItems()
	AddSystrayMenuItems(items []CallbackInfo)

	// destroy gives implementation a chance to
	// clean up resources
	destroy()
}

func New(iconPath string, clientPath string) *Systray {
	// Wrap the platform-specific implementation in a
	// public concrete type
	st := &Systray{_NewSystray(iconPath, clientPath)}

	// Track the cleanup of the public Systray instance,
	// so that we can decref the wrapped private instance
	runtime.SetFinalizer(st, func(ptr *Systray) {
		ptr.destroy()
	})

	return st
}

type CallbackInfo struct {
	ItemName string
	Callback func()
	Disabled bool
	Checked  bool
}
