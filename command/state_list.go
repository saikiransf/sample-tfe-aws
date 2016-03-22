package command

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/cli"
)

// StateListCommand is a Command implementation that lists the resources
// within a state file.
type StateListCommand struct {
	Meta
}

func (c *StateListCommand) Run(args []string) int {
	args = c.Meta.process(args, true)

	cmdFlags := c.Meta.flagSet("state list")
	cmdFlags.StringVar(&c.Meta.statePath, "state", DefaultStateFilename, "path")
	cmdFlags.StringVar(&c.Meta.backupPath, "backup", "", "path")
	if err := cmdFlags.Parse(args); err != nil {
		return cli.RunResultHelp
	}

	args = cmdFlags.Args()
	if len(args) > 1 {
		c.Ui.Error(
			"At most one argument expected: the pattern to list\n")
		return cli.RunResultHelp
	}

	state, err := c.State()
	if err != nil {
		c.Ui.Error(fmt.Sprintf(errStateLoadingState, err))
		return cli.RunResultHelp
	}

	filter := &terraform.StateFilter{State: state.State()}
	println(filter)

	return 0
}

func (c *StateListCommand) Help() string {
	helpText := `
Usage: terraform state list [options] [pattern]

  List resources in the Terraform state.

  This command lists resources in the Terraform state. The pattern argument
  can be used to filter the resources by resource or module. If no pattern
  is given, all resources are listed.

  The pattern argument is meant to provide very simple filtering. For
  advanced filtering, please use tools such as "grep". The output of this
  command is designed to be friendly for this usage.

  The pattern argument accepts any resource targeting syntax. Please
  refer to the documentation on resource targeting syntax for more
  information.

Options:

  -backup=path        Path to backup the existing state file before
                      modifying. Defaults to the the input state path
                      plus a timestamp with the ".backup" extension.
                      Backups cannot be disabled for state management commands.

  -state=statefile    Path to a Terraform state file to use to look
                      up Terraform-managed resources. By default it will
                      use the state "terraform.tfstate" if it exists.

`
	return strings.TrimSpace(helpText)
}

func (c *StateListCommand) Synopsis() string {
	return "List resources in the state"
}

const errStateLoadingState = `Error loading the state: %[1]s

Please ensure that your Terraform state exists and that you've
configured it properly. You can use the "-state" flag to point
Terraform at another state file.`
