package resources

import (
	"embed"
	"io/fs"
)

const (
	Configs   = "configs"
	Scripts   = "scripts"
	Binaries  = "binaries"
	Terraform = "terraform"
)

//go:embed terraform
var terraformResource embed.FS

func SubFolder(filesystem fs.FS, subfolder string) fs.FS {
	result, err := fs.Sub(filesystem, subfolder)
	if err != nil {
		return embed.FS{}
	}
	return result
}

func DefaultTerraformResources() fs.FS {
	return SubFolder(terraformResource, "terraform")
}
