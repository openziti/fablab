package zitilib_examples

import "github.com/netfoundry/fablab/kernel/model"

var modelScope = model.Scope{
	Variables: model.Variables{
		"zitilib": model.Variables{
			"fabric": model.Variables{
				"data_plane_protocol": &model.Variable{Default: "tls"},
			},
		},
		"environment": &model.Variable{Required: true},
		"credentials": model.Variables{
			"aws": model.Variables{
				"access_key":   &model.Variable{Required: true, Sensitive: true},
				"secret_key":   &model.Variable{Required: true, Sensitive: true},
				"ssh_key_name": &model.Variable{Required: true},
			},
			"ssh": model.Variables{
				"key_path": &model.Variable{Required: true},
				"username": &model.Variable{Default: "fedora"},
			},
		},
		"distribution": model.Variables{
			"rsync_bin": &model.Variable{Default: "rsync"},
			"ssh_bin":   &model.Variable{Default: "ssh"},
		},
		"instance_type": &model.Variable{Default: "t2.micro"},
	},
}
