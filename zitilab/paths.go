package zitilab

import "path/filepath"

func ZitiRoot() string {
	return zitiRoot
}

func zitiBinaries() string {
	return filepath.Join(zitiRoot, "bin")
}

func ZitiCli() string {
	return filepath.Join(zitiBinaries(), "ziti")
}

func ZitiFabricCli() string {
	return filepath.Join(zitiBinaries(), "ziti-fabric")
}

var zitiRoot string