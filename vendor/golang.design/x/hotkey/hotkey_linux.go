// Copyright 2021 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.
//
// Written by Changkun Ou <changkun.de>

//go:build linux

package hotkey

/*
#cgo LDFLAGS: -lX11

#include <stdint.h>

int displayTest();
int waitHotkey(uintptr_t hkhandle, unsigned int mod, int key);
*/
import "C"
import (
	"context"
	"errors"
	"runtime"
	"runtime/cgo"
	"sync"
)

const errmsg = `Failed to initialize the X11 display, and the clipboard package
will not work properly. Install the following dependency may help:

	apt install -y libx11-dev
If the clipboard package is in an environment without a frame buffer,
such as a cloud server, it may also be necessary to install xvfb:
	apt install -y xvfb
and initialize a virtual frame buffer:
	Xvfb :99 -screen 0 1024x768x24 > /dev/null 2>&1 &
	export DISPLAY=:99.0
Then this package should be ready to use.
`

func init() {
	if C.displayTest() != 0 {
		panic(errmsg)
	}
}

type platformHotkey struct {
	mu         sync.Mutex
	registered bool
	ctx        context.Context
	cancel     context.CancelFunc
	canceled   chan struct{}
}

// Nothing needs to do for register
func (hk *Hotkey) register() error {
	hk.mu.Lock()
	if hk.registered {
		hk.mu.Unlock()
		return errors.New("hotkey already registered.")
	}
	hk.registered = true
	hk.ctx, hk.cancel = context.WithCancel(context.Background())
	hk.canceled = make(chan struct{})
	hk.mu.Unlock()

	go hk.handle()
	return nil
}

// Nothing needs to do for unregister
func (hk *Hotkey) unregister() error {
	hk.mu.Lock()
	defer hk.mu.Unlock()
	if !hk.registered {
		return errors.New("hotkey is not registered.")
	}
	hk.cancel()
	hk.registered = false
	<-hk.canceled
	return nil
}

// handle registers an application global hotkey to the system,
// and returns a channel that will signal if the hotkey is triggered.
func (hk *Hotkey) handle() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// KNOWN ISSUE: if a hotkey is grabbed by others, C side will crash the program

	var mod Modifier
	for _, m := range hk.mods {
		mod = mod | m
	}
	h := cgo.NewHandle(hk)
	defer h.Delete()

	for {
		select {
		case <-hk.ctx.Done():
			close(hk.canceled)
			return
		default:
			_ = C.waitHotkey(C.uintptr_t(h), C.uint(mod), C.int(hk.key))
		}
	}
}

//export hotkeyDown
func hotkeyDown(h uintptr) {
	hk := cgo.Handle(h).Value().(*Hotkey)
	hk.keydownIn <- Event{}
}

//export hotkeyUp
func hotkeyUp(h uintptr) {
	hk := cgo.Handle(h).Value().(*Hotkey)
	hk.keyupIn <- Event{}
}

// Modifier represents a modifier.
type Modifier uint32

// All kinds of Modifiers
// See /usr/include/X11/X.h
const (
	ModCtrl  Modifier = (1 << 2)
	ModShift Modifier = (1 << 0)
	Mod1     Modifier = (1 << 3)
	Mod2     Modifier = (1 << 4)
	Mod3     Modifier = (1 << 5)
	Mod4     Modifier = (1 << 6)
	Mod5     Modifier = (1 << 7)
)

// Key represents a key.
// See /usr/include/X11/keysymdef.h
type Key uint16

// All kinds of keys
const (
	KeySpace Key = 0x0020
	Key1     Key = 0x0030
	Key2     Key = 0x0031
	Key3     Key = 0x0032
	Key4     Key = 0x0033
	Key5     Key = 0x0034
	Key6     Key = 0x0035
	Key7     Key = 0x0036
	Key8     Key = 0x0037
	Key9     Key = 0x0038
	Key0     Key = 0x0039
	KeyA     Key = 0x0061
	KeyB     Key = 0x0062
	KeyC     Key = 0x0063
	KeyD     Key = 0x0064
	KeyE     Key = 0x0065
	KeyF     Key = 0x0066
	KeyG     Key = 0x0067
	KeyH     Key = 0x0068
	KeyI     Key = 0x0069
	KeyJ     Key = 0x006a
	KeyK     Key = 0x006b
	KeyL     Key = 0x006c
	KeyM     Key = 0x006d
	KeyN     Key = 0x006e
	KeyO     Key = 0x006f
	KeyP     Key = 0x0070
	KeyQ     Key = 0x0071
	KeyR     Key = 0x0072
	KeyS     Key = 0x0073
	KeyT     Key = 0x0074
	KeyU     Key = 0x0075
	KeyV     Key = 0x0076
	KeyW     Key = 0x0077
	KeyX     Key = 0x0078
	KeyY     Key = 0x0079
	KeyZ     Key = 0x007a

	KeyReturn Key = 0xff0d
	KeyEscape Key = 0xff1b
	KeyDelete Key = 0xffff
	KeyTab    Key = 0xff1b

	KeyLeft  Key = 0xff51
	KeyRight Key = 0xff53
	KeyUp    Key = 0xff52
	KeyDown  Key = 0xff54

	KeyF1  Key = 0xffbe
	KeyF2  Key = 0xffbf
	KeyF3  Key = 0xffc0
	KeyF4  Key = 0xffc1
	KeyF5  Key = 0xffc2
	KeyF6  Key = 0xffc3
	KeyF7  Key = 0xffc4
	KeyF8  Key = 0xffc5
	KeyF9  Key = 0xffc6
	KeyF10 Key = 0xffc7
	KeyF11 Key = 0xffc8
	KeyF12 Key = 0xffc9
	KeyF13 Key = 0xffca
	KeyF14 Key = 0xffcb
	KeyF15 Key = 0xffcc
	KeyF16 Key = 0xffcd
	KeyF17 Key = 0xffce
	KeyF18 Key = 0xffcf
	KeyF19 Key = 0xffd0
	KeyF20 Key = 0xffd1
)
