package cli

import "github.com/netfoundry/fablab/kernel/internal"

func Figlet(text string) {
	internal.Figlet(text)
}

func FigletMini(text string) {
	internal.FigletMini(text)
}