package util

import (
	"path"
	"runtime"
)

/*
	TODO:
	Part of me feels this will fail...

	If the binary is distributed separate from the source files, how does it know where the source files are at runtime?
	^ Especially if the filepaths are embedded at time of build...
*/

// Basically a port of __dirname from javascript
func DirName() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
