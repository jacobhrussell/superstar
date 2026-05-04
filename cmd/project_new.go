package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	projectNewDir   string
	projectNewAgent string
)

var projectNewCmd = &cobra.Command{
	Use:     "new [name]",
	Aliases: []string{"n"},
	Short:   "Create a new project entry in the config",
	Args:    cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if isInteractive(args) {
			return nil
		}
		if projectNewDir == "" {
			return errors.New("--dir is required")
		}
		if projectNewAgent == "" {
			return errors.New("--agent is required")
		}
		return validateAgent(projectNewAgent)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var name, dir, agent string

		if isInteractive(args) {
			var err error
			name, dir, agent, err = promptProject(cmd.InOrStdin(), cmd.OutOrStdout())
			if err != nil {
				return err
			}
		} else {
			name = args[0]
			dir = projectNewDir
			agent = projectNewAgent
		}

		if name == "" {
			return errors.New("project name cannot be empty")
		}

		key := "projects." + name
		if viper.IsSet(key) {
			return fmt.Errorf("project %q already exists", name)
		}

		cfgDir, err := configDir()
		if err != nil {
			return fmt.Errorf("could not resolve home directory: %w", err)
		}
		path, err := configPath()
		if err != nil {
			return fmt.Errorf("could not resolve home directory: %w", err)
		}
		if err := os.MkdirAll(cfgDir, 0o755); err != nil {
			return fmt.Errorf("could not create config directory: %w", err)
		}

		viper.Set(key+".dir", dir)
		viper.Set(key+".agent", agent)

		viper.SetConfigFile(path)
		if err := viper.WriteConfig(); err != nil {
			return fmt.Errorf("could not write config: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "created project %q in %s\n", name, path)
		return nil
	},
}

func isInteractive(args []string) bool {
	return len(args) == 0 && projectNewDir == "" && projectNewAgent == ""
}

func promptProject(in io.Reader, out io.Writer) (name, dir, agent string, err error) {
	reader := bufio.NewReader(in)

	for {
		if _, err = fmt.Fprint(out, "Project name: "); err != nil {
			return
		}
		name, err = readLine(reader)
		if err != nil {
			return
		}
		if name == "" {
			err = errors.New("project name cannot be empty")
			return
		}
		if !viper.IsSet("projects." + name) {
			break
		}
		fmt.Fprintf(out, "project %q already exists, try a different name\n", name)
	}

	fmt.Fprint(out, "Directory: ")
	dir, err = readLine(reader)
	if err != nil {
		return
	}
	if dir == "" {
		err = errors.New("directory cannot be empty")
		return
	}

	for {
		fmt.Fprintf(out, "Agent (%s): ", strings.Join(validAgents, "/"))
		agent, err = readLine(reader)
		if err != nil {
			return
		}
		if agent == "" {
			err = errors.New("agent cannot be empty")
			return
		}
		if validateAgent(agent) == nil {
			break
		}
		fmt.Fprintln(out, "invalid agent, try again")
	}

	return
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func init() {
	projectNewCmd.Flags().StringVarP(&projectNewDir, "dir", "d", "", "directory for the project")
	projectNewCmd.Flags().StringVarP(&projectNewAgent, "agent", "a", "", fmt.Sprintf("agent for the project (%s)", strings.Join(validAgents, ", ")))
	projectCmd.AddCommand(projectNewCmd)
}
