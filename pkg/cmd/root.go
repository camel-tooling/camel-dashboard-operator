/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"

	"github.com/squakez/camel-dashboard-operator/pkg/client"
)

// RootCmdOptions --.
//
//nolint:containedctx
type RootCmdOptions struct {
	RootContext   context.Context    `mapstructure:"-"`
	Context       context.Context    `mapstructure:"-"`
	ContextCancel context.CancelFunc `mapstructure:"-"`
	_client       client.Client      `mapstructure:"-"`
	Flags         *viper.Viper       `mapstructure:"-"`
	KubeConfig    string             `mapstructure:"kube-config"`
	Namespace     string             `mapstructure:"namespace"`
	Verbose       bool               `mapstructure:"verbose" yaml:",omitempty"`
}

// NewKamelCommand --.
func NewKamelCommand(ctx context.Context) (*cobra.Command, error) {
	childCtx, childCancel := context.WithCancel(ctx)
	options := RootCmdOptions{
		RootContext:   ctx,
		Context:       childCtx,
		ContextCancel: childCancel,
		Flags:         viper.New(),
	}

	cmd := kamelPreAddCommandInit(&options)
	addKamelSubcommands(cmd, &options)

	if err := addHelpSubCommands(cmd); err != nil {
		return cmd, err
	}

	err := kamelPostAddCommandInit(cmd, options.Flags)

	return cmd, err
}

func kamelPreAddCommandInit(options *RootCmdOptions) *cobra.Command {
	cmd := cobra.Command{
		PersistentPreRunE: options.preRun,
		Use:               "camel-dashboard",
		Short:             "camel-dashboard",
		Long:              "camel-dashboard",
		SilenceUsage:      true,
	}

	cmd.PersistentFlags().StringVar(&options.KubeConfig, "kube-config", os.Getenv("KUBECONFIG"), "Path to the kube config file to use for CLI requests")
	cmd.PersistentFlags().StringVarP(&options.Namespace, "namespace", "n", "", "Namespace to use for all operations")
	cmd.PersistentFlags().BoolVarP(&options.Verbose, "verbose", "V", false, "Verbose logging")

	cobra.AddTemplateFunc("wrappedFlagUsages", wrappedFlagUsages)
	cmd.SetUsageTemplate(usageTemplate)

	return &cmd
}

func kamelPostAddCommandInit(cmd *cobra.Command, v *viper.Viper) error {
	configName := os.Getenv("KAMEL_CONFIG_NAME")
	if configName == "" {
		configName = DefaultConfigName
	}

	v.SetConfigName(configName)

	configPath := os.Getenv("KAMEL_CONFIG_PATH")
	if configPath != "" {
		// if a specific config path is set, don't add
		// default locations
		v.AddConfigPath(configPath)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath(".kamel")
		v.AddConfigPath("$HOME/.kamel")
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(
		".", "_",
		"-", "_",
	))

	if err := v.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return err
		}
	}

	return nil
}

func addKamelSubcommands(cmd *cobra.Command, options *RootCmdOptions) {
	cmd.AddCommand(cmdOnly(newCmdOperator(options)))
}

func addHelpSubCommands(cmd *cobra.Command) error {
	cmd.InitDefaultHelpCmd()

	var helpCmd *cobra.Command
	for _, c := range cmd.Commands() {
		if c.Name() == "help" {
			helpCmd = c
			break
		}
	}

	if helpCmd == nil {
		return errors.New("could not find any configured help command")
	}

	return nil
}

func (command *RootCmdOptions) preRun(cmd *cobra.Command, _ []string) error {
	c, err := command.GetCmdClient()
	if err != nil {
		return fmt.Errorf("cannot get command client: %w", err)
	}
	if command.Namespace == "" {
		current := command.Flags.GetString("kamel.config.default-namespace")
		if current == "" {
			defaultNS, err := c.GetCurrentNamespace(command.KubeConfig)
			if err != nil {
				return fmt.Errorf("cannot get current namespace: %w", err)
			}
			current = defaultNS
		}
		err = cmd.Flag("namespace").Value.Set(current)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetCmdClient returns the client that can be used from command line tools.
func (command *RootCmdOptions) GetCmdClient() (client.Client, error) {
	// Get the pre-computed client
	if command._client != nil {
		return command._client, nil
	}
	var err error
	command._client, err = command.NewCmdClient()
	return command._client, err
}

// NewCmdClient returns a new client that can be used from command line tools.
func (command *RootCmdOptions) NewCmdClient() (client.Client, error) {
	return client.NewOutOfClusterClient(command.KubeConfig)
}

func wrappedFlagUsages(cmd *cobra.Command) string {
	width := 80
	if w, _, err := term.GetSize(0); err == nil {
		width = w
	}
	return cmd.Flags().FlagUsagesWrapped(width - 1)
}

var usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{ wrappedFlagUsages . | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
