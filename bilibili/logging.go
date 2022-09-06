/*
This file defines the common struct of logger pointers used in modules of this package.
*/
package bilibili

import "log"

type loggerCommon struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}
