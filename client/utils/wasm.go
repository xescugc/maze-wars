package utils

import "runtime"

func IsWASM() bool { return runtime.GOOS == "js" && runtime.GOARCH == "wasm" }
