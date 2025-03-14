// Package commands ...
package commands

import (
	"path/filepath"
	"strings"

	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"

	"github.com/nanobox-io/nanobox/commands/registry"
	"github.com/nanobox-io/nanobox/util/config"
	"github.com/nanobox-io/nanobox/util/display"
	"github.com/nanobox-io/nanobox/util/mixpanel"
	"github.com/nanobox-io/nanobox/util/update"
)

//
var (
	// debug mode
	debugMode bool

	// display level debug
	displayDebugMode bool

	// display level trace
	displayTraceMode bool

	//
	internalCommand bool

	//
	endpoint string

	// NanoboxCmd ...
	NanoboxCmd = &cobra.Command{
		Use:   "nanobox",
		Short: "",
		Long:  ``,
		PersistentPreRun: func(ccmd *cobra.Command, args []string) {

			// report the command usage to mixpanel
			mixpanel.Report(strings.Replace(ccmd.CommandPath(), "nanobox ", "", 1))

			// alert the user if an update is needed
			update.Check()

			// TODO: look into global messaging
			if internalCommand {
				registry.Set("internal", internalCommand)
				// setup a file logger, this will be replaced in verbose mode.
				fileLogger, _ := lumber.NewAppendLogger(filepath.ToSlash(filepath.Join(config.GlobalDir(), "nanobox.log")))
				lumber.SetLogger(fileLogger)

			}

			if endpoint != "" {
				registry.Set("endpoint", endpoint)
			}

			registry.Set("debug", debugMode)

			// setup the display output
			if displayDebugMode {
				lumber.Level(lumber.DEBUG)
				display.Summary = false
				display.Level = "debug"
			}

			if displayTraceMode {
				lumber.Level(lumber.TRACE)
				display.Summary = false
				display.Level = "trace"
			}

		},

		//
		Run: func(ccmd *cobra.Command, args []string) {

			// fall back on default help if no args/flags are passed
			ccmd.HelpFunc()(ccmd, args)
		},
	}
)

// init creates the list of available nanobox commands and sub commands
func init() {

	// persistent flags
	NanoboxCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "", "production endpoint")
	NanoboxCmd.PersistentFlags().MarkHidden("endpoint")
	NanoboxCmd.PersistentFlags().BoolVarP(&internalCommand, "internal", "", false, "Skip pre-requisite checks")
	NanoboxCmd.PersistentFlags().MarkHidden("internal")
	NanoboxCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "", false, "In the event of a failure, drop into debug context")
	NanoboxCmd.PersistentFlags().BoolVarP(&displayDebugMode, "verbose", "v", false, "Increases display output and sets level to debug")
	NanoboxCmd.PersistentFlags().BoolVarP(&displayTraceMode, "trace", "t", false, "Increases display output and sets level to trace")

	// subcommands
	NanoboxCmd.AddCommand(ConfigureCmd)
	NanoboxCmd.AddCommand(RunCmd)
	NanoboxCmd.AddCommand(BuildCmd)
	NanoboxCmd.AddCommand(CompileCmd)
	NanoboxCmd.AddCommand(DeployCmd)
	NanoboxCmd.AddCommand(ConsoleCmd)
	NanoboxCmd.AddCommand(RemoteCmd)
	NanoboxCmd.AddCommand(StatusCmd)
	NanoboxCmd.AddCommand(LoginCmd)
	NanoboxCmd.AddCommand(LogoutCmd)
	NanoboxCmd.AddCommand(CleanCmd)
	NanoboxCmd.AddCommand(InfoCmd)
	NanoboxCmd.AddCommand(TunnelCmd)
	NanoboxCmd.AddCommand(ImplodeCmd)
	NanoboxCmd.AddCommand(DestroyCmd)
	NanoboxCmd.AddCommand(StartCmd)
	NanoboxCmd.AddCommand(StopCmd)
	NanoboxCmd.AddCommand(UpdateCmd)
	NanoboxCmd.AddCommand(EvarCmd)
	NanoboxCmd.AddCommand(DnsCmd)
	NanoboxCmd.AddCommand(LogCmd)
	NanoboxCmd.AddCommand(VersionCmd)

	// hidden subcommands
	NanoboxCmd.AddCommand(EnvCmd)
	NanoboxCmd.AddCommand(InspectCmd)
}
