#import <Cocoa/Cocoa.h>
#include <strings.h>
#include "_cgo_export.h"

typedef void(*CallbackFunc)(void *, GoInt);
typedef struct {
    void *manager;      // The Go object that will handle the click
    unsigned int index; // Representative index of which go-side callback to invoke
    CallbackFunc callback;
    bool enabled;
    bool checked;
} MenuCallbackInfo;

@interface AppDelegate: NSObject <NSApplicationDelegate>
{
    NSMutableDictionary *icons;
    NSStatusItem *_statusItem;
}
- (void)applicationDidFinishLaunching:(NSNotification *)aNotification;
- (void)showIcon:(NSString*)path hint:(NSString *)hint;
- (void)addMenuItem:(NSString*)item manager:(void*)manager index:(int)index enabled:(bool)enabled checked:(bool)checked callback:(CallbackFunc)callback;
- (void)clearMenuItems;

- (IBAction)clicked:(id)sender;
- (IBAction)menuItem:(id)sender;


@property (readwrite, retain) NSStatusItem *statusItem;
@property (assign) IBOutlet NSWindow *window;

@end

@implementation AppDelegate
NSMenu *m_menu;
void *m_manager;
const char *m_initialIconPath;
const char *m_initialHint;

@synthesize window = _window;
@synthesize statusItem = _statusItem;

- (id)init:(void *)manager iconPath:(const char *)iconPath hint:(const char*)hint{
    if ((self = [super init])) {
        m_manager = manager;
        m_initialIconPath = iconPath;
        m_initialHint = hint;
    }
    return self;
}

- (void)dealloc {
    [_statusItem release]; _statusItem = nil;
    [super dealloc];
}

- (void)applicationDidFinishLaunching:(NSNotification *)aNotification
{
    NSLog(@"Setting up menubar");

    // Create a menubar item
    self.statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];

    // Useful for debugging if icon loading is broken (e.g., icons don't exist)
    [self.statusItem setTitle:@"SystrayTest"];

    // Set up a general click handler - this will happen in addition to any menu
    [self.statusItem setAction:@selector(clicked:)];

    // Create our menu and add some items
    //NSMenu *statusMenu = [[[NSMenu allocWithZone:[NSMenu menuZone]] initWithTitle:@"Custom"] autorelease];
    NSMenu *statusMenu = [[[NSMenu alloc] initWithTitle:@"Custom"] autorelease];
    [statusMenu setAutoenablesItems:NO];
    
    // TODO: the app is now solely responsible for managing termination, so we don't add the
    // default Quit item - consider whether or not a placeholder or label item should be added
    // in its stead
    [self.statusItem setMenu:statusMenu];

    // Set up an icon cache for loaded image data
    icons = [[NSMutableDictionary alloc] init];

    // Set our initial icon and hint/tooltip
    NSString *nsIcon = [NSString stringWithCString:m_initialIconPath encoding:NSASCIIStringEncoding];
    NSString *nsHint = [NSString stringWithCString:m_initialHint encoding:NSASCIIStringEncoding];
    [self showIcon:nsIcon hint:nsHint];

    // Tell our caller that the menubar item has been created
    NSLog(@"AppDelegate setup finished, calling menuCreatedCallback");
    menuCreatedCallback(m_manager);
}

// Add a new menu item, with callback and metadata
- (void)addMenuItem:(NSString*)item manager:(void*)manager index:(int)index enabled:(bool)enabled checked:(bool)checked callback:(CallbackFunc)callback {
    MenuCallbackInfo callbackInfo;
    callbackInfo.manager = manager;
    callbackInfo.index = index;
    callbackInfo.callback = callback;
    callbackInfo.enabled = enabled;
    callbackInfo.checked = checked;
    //NSMenuItem *newItem = [[[NSMenuItem allocWithZone:[NSMenu menuZone]] initWithTitle:item action:@selector(menuItem:) keyEquivalent:@""] autorelease];
    NSMenuItem *newItem = [[[NSMenuItem alloc] initWithTitle:item action:@selector(menuItem:) keyEquivalent:@""] autorelease];
    if (enabled) {
        [newItem setEnabled:YES];
    } else {
        [newItem setEnabled:NO];
    }
    if (checked) {
        [newItem setState:NSOnState];
    }
    [newItem setRepresentedObject:[NSValue value:&callbackInfo withObjCType:@encode(MenuCallbackInfo)]];
    
    [[self.statusItem menu] addItem:newItem];
}

- (void)clearMenuItems {
    [[self.statusItem menu] removeAllItems];
}

// Process a previously added menu item, extracting the callback info and invoking
// the callback
- (IBAction)menuItem:(id)sender {
    NSLog(@"Menu item!");
    NSValue *nsCallbackInfo = [sender representedObject];
    MenuCallbackInfo callbackInfo;
    [nsCallbackInfo getValue:&callbackInfo];
    if (callbackInfo.manager != nil && callbackInfo.callback != nil) {
        NSLog(@"Issuing callback");
        callbackInfo.callback(callbackInfo.manager, callbackInfo.index);
    }
}

// Provide any necessary response to a click in addition to the menu
- (IBAction)clicked:(id)sender {
    // This could include a generic callback over the fence to Go
    // if we want to do something like reset an idle timer.
    // Fun fact: this does not get called if a menu is set, so utility is limited
    //NSLog(@"clicked");
}

// Set the menubar icon and hint, if any. Either value may
// be nil, in which case no action is taken.
// TBD: decide if nil should mean "remove" instead
- (void)showIcon:(NSString*)path hint:(NSString *)hint {
    if (path) {
        NSImage *icon = [icons objectForKey:path];
        if (!icon) {
            NSLog(@"Creating new icon from file");
            icon = [[[NSImage alloc] initWithContentsOfFile:path] autorelease];
            if (icon) {
                [icons setValue:icon forKey:path];
            }
        }
        if (icon) {
            NSLog(@"Setting icon image");
            [self.statusItem setImage:icon];
        }
    }
    if (hint) {
        [self.statusItem setToolTip:hint];
    }
}
@end

// External API - these are the C functions that are directly callable
// from Go.

// Run the cocoa application's main event loop. This must be run on the main
// thread and will block, so architect accordingly, especially cross-platform.

// TODO: Enforce the main thread restriction earlier, rather than letting some
// NS code assert
// TODO: Consider breaking this into separate setup and run functions, so the app
// can be stopped and restarted
void runApplication(const char *title,
                    const char *initialIcon,
                    const char *initialHint,
                    void *manager) {
    [NSAutoreleasePool new];
    [NSApplication sharedApplication];
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];

    AppDelegate *delegate = [[[AppDelegate alloc] init:manager iconPath:initialIcon hint:initialHint] autorelease];
    [NSApp setDelegate:delegate];

    //[NSApp activateIgnoringOtherApps:YES];
    NSLog(@"Running main application");
    [NSApp run];
}

// Stop the application
void stopApplication(void) {
    [NSApplication sharedApplication];
    [NSApp stop:nil];
}

// Set the currently displayed icon
// TODO: figure out how we want to pass unicode
void setIcon(const char *path) {
    NSString *nsPath = [NSString stringWithCString:path encoding:NSASCIIStringEncoding];
    [[NSApp delegate] showIcon:nsPath hint:nil];
}

// Set the currently displayed hint
void setHint(const char *hint) {
    [NSAutoreleasePool new];
    NSString *nsHint = [NSString stringWithCString:hint encoding:NSASCIIStringEncoding];
    [[NSApp delegate] showIcon:nil hint:nsHint];
}

// Add a new item to the menu, with some (opaque) info on how to process it back on
// the other side
void addSystrayMenuItem(const char *item, void *object, unsigned int index, unsigned char enabled, unsigned char checked) {
    [NSAutoreleasePool new];
    [[NSApp delegate] addMenuItem:[NSString stringWithCString:item encoding:NSASCIIStringEncoding]
                                   manager:object
                                   index:index
                                   enabled:enabled
                                   checked:checked
                                   callback:&menuClickCallback];
}

// Clear all existing items from the menu
void clearSystrayMenuItems(void) {
    [NSAutoreleasePool new];
    [[NSApp delegate] clearMenuItems];
}

