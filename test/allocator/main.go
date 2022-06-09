/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package main

import (
	"log"
	"os"

	"github.com/go-logr/stdr"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/app"
)

func main() {
	logger := stdr.NewWithOptions(
		log.New(os.Stdout, "", log.LstdFlags),
		stdr.Options{LogCaller: stdr.All},
	)

	a := app.NewApp(logger, app.NewFactory(logger))
	a.Run()
}
