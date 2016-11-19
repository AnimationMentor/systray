Modified to directly invoke objc via cgo for darwin. This adds a tray_darwin.c source file
and blends the original systray.Client standalone app functionality with that of the
cgo-compiled integrated objc from github.com/cratonica/trayhost (which did a dock icon/menu
for darwin and required cgo for all platforms).

This carries two important ramifications:
1) Darwin app (and linux, eventually) must be compiled on target platform
   (or cross-compiled for cgo, when/if that happens officially)
2) Systray.Run() *must* be executed on the main thread in order to play 
   nice with Cocoa

Also includes support for adding menu items with Go callbacks.

[================== original content below ======================]

Systray (Trayicon/Menu Extras)
=======


## Cross platform systray for golang

Instead of gui program, "go-program + systray + web-console" might be a interesting choise.


# Platform

Mac: avalid  
Win: avalid  
Linux: coming soon  


## Run example

Mac:
```
cd example
go run icons/mac systray
```

## How it works

Win:  
    [your code in go] -> [systray: win32 api call in go]

Mac:  
    [your code in go] -> [systray.Server in go] -(tcp)-> [systray.Client in objc]

Linux:  
    [your code in go] -> [systray.Server in go] -(tcp)-> [systray.Client in c]


