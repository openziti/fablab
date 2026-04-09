/*
	(c) Copyright NetFoundry Inc. Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package subcmd

import (
	"os"
	"path/filepath"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Execute() error {
	return RootCmd.Execute()
}

func init() {
	options := pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti/").NoColor()
	pfxlog.GlobalInit(logrus.InfoLevel, options)

	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	RootCmd.PersistentFlags().StringVarP(&model.CliInstanceId, "instance", "i", "", "specify the instance to use")
	RootCmd.PersistentFlags().StringVar(&logFormatter, "log-formatter", "", "Specify log formatter [json|pfxlog|text]")
	RootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Tee log output to the specified file")
}

var RootCmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "The Fabulous Laboratory",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		switch logFormatter {
		case "pfxlog":
			options := pfxlog.DefaultOptions().StartingToday()
			logrus.SetFormatter(pfxlog.NewFormatter(options))
		case "json":
			logrus.SetFormatter(&logrus.JSONFormatter{})
		case "text":
			logrus.SetFormatter(&logrus.TextFormatter{})
		default:
			// let logrus do its own thing
		}

		if logFile != "" {
			f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				logrus.WithError(err).Fatalf("failed to open log file %q", logFile)
			}
			logrus.AddHook(&fileLogHook{file: f})
		}
	},
}

var verbose bool
var logFormatter string
var logFile string

// fileLogHook is a logrus hook that writes formatted log entries to a file.
// It works alongside the TUI hook, which redirects logrus output to discard.
type fileLogHook struct {
	file *os.File
}

func (h *fileLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *fileLogHook) Fire(entry *logrus.Entry) error {
	formatted, err := logrus.StandardLogger().Formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = h.file.Write(formatted)
	return err
}
