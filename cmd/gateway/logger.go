package main

import logger "github.com/sconklin/go-logger"

// You can manage verbosity of log output
// in the package by changing last parameter value.
var log = logger.NewPackageLogger("main",
	// logger.DebugLevel,
	logger.InfoLevel,
)
