/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package model

import "os"

func GetEnvVarWithDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		val = defaultVal
	}
	return val
}
