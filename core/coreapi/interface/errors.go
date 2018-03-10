package iface

import "errors"

var ErrIsDir = errors.New("object is a directory")
var ErrOffline = errors.New("this action must be run in online mode, try running 'ipfs daemon' first")
