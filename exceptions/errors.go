package exceptions

import "errors"

var SilenceRequestedError = errors.New("prompt is disabled because user has requested silence")
var NoReleaseFoundError = errors.New("could not find any releases")
