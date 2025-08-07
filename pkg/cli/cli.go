package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//
// CLI
//

type CLI struct {
	args []string
	
	// Only store chdir since we use it immediately
	chdir string
}

type Command interface {
	Name() string
	Description() string
	Run(args []string) error
}

func New(args []string) *CLI {
	return &CLI{
		args: args,
		// Initialize with defaults
		chdir: getEnvOrDefault("MONAD_CHDIR", "."),
	}
}

func (c *CLI) Run() error {
	// Parse global options first
	remainingArgs, err := c.parseGlobalOptions(c.args[1:])
	if err != nil {
		return err
	}

	if len(remainingArgs) == 0 {
		c.showHelp()
		return nil
	}

	// Change directory if specified
	if c.chdir != "." {
		if err := os.Chdir(c.chdir); err != nil {
			return fmt.Errorf("failed to change directory to %s: %w", c.chdir, err)
		}
	}

	subcommand := remainingArgs[0]
	subArgs := remainingArgs[1:]

	switch subcommand {
	case "deploy":
		return c.runDeploy(subArgs)
	case "destroy":
		return c.runDestroy(subArgs)
	case "list":
		return c.runList(subArgs)
	case "ecr":
		return c.runEcr(subArgs)
	case "init":
		return c.runInit(subArgs)
	case "data":
		return c.runData(subArgs)
	case "help", "-h", "--help":
		c.showHelp()
		return nil
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", subcommand)
		c.showHelp()
		return fmt.Errorf("unknown command: %s", subcommand)
	}
}

func (c *CLI) showHelp() {
	basisHelp, err := GenerateHelpText(BasisConfig{})
	if err != nil {
		basisHelp = "Error generating help text"
	}
	
	fmt.Printf(`Usage: monad [options] [basis] <command> [options]

Commands:
  deploy                 deploy a service
  destroy                destroy a service
  list                   list services
  ecr                    ecr commands
  init                   initialize a service
  data                   contextual template data

Options:
  --chdir path           change working directory [default: .]
  --help, -h             display this help and exit

Basis:
%s

Use "monad <command> -h" for more information about a command.
`, basisHelp)
}

//
// Subcommand implementations
//

func (c *CLI) runDeploy(args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(`Usage: monad deploy [options]

Deploy a service to AWS using basis and config.

Options:
  -h, --help             show this help message
`)
		return nil
	}

	fmt.Println("Running deploy command...")
	// TODO: Implement deploy logic
	return fmt.Errorf("deploy not yet implemented")
}

func (c *CLI) runDestroy(args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(`Usage: monad destroy [options]

Destroy a deployed service from AWS.

Options:
  -h, --help             show this help message
`)
		return nil
	}

	fmt.Println("Running destroy command...")
	// TODO: Implement destroy logic
	return fmt.Errorf("destroy not yet implemented")
}

func (c *CLI) runList(args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(`Usage: monad list [options]

List deployed services.

Options:
  -h, --help             show this help message
`)
		return nil
	}

	fmt.Println("Running list command...")
	// TODO: Implement list logic
	return fmt.Errorf("list not yet implemented")
}

func (c *CLI) runEcr(args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(`Usage: monad ecr <subcommand> [options]

ECR container registry commands.

Subcommands:
  login                  login to ECR registry
  push                   push image to registry
  pull                   pull image from registry

Options:
  -h, --help             show this help message
`)
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("ecr command requires a subcommand")
	}

	ecrCommand := args[0]
	ecrArgs := args[1:]

	switch ecrCommand {
	case "login":
		return c.runEcrLogin(ecrArgs)
	case "push":
		return c.runEcrPush(ecrArgs)
	case "pull":
		return c.runEcrPull(ecrArgs)
	default:
		return fmt.Errorf("unknown ecr command: %s", ecrCommand)
	}
}

func (c *CLI) runEcrLogin(args []string) error {
	fmt.Println("Running ECR login...")
	// TODO: Implement ECR login logic
	return fmt.Errorf("ecr login not yet implemented")
}

func (c *CLI) runEcrPush(args []string) error {
	fmt.Println("Running ECR push...")
	// TODO: Implement ECR push logic
	return fmt.Errorf("ecr push not yet implemented")
}

func (c *CLI) runEcrPull(args []string) error {
	fmt.Println("Running ECR pull...")
	// TODO: Implement ECR pull logic
	return fmt.Errorf("ecr pull not yet implemented")
}

func (c *CLI) runInit(args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(`Usage: monad init [language] [options]

Initialize a new service with scaffolding.

Arguments:
  language               programming language (go, node, python, ruby, shell)

Options:
  -h, --help             show this help message
  --policy               include policy template
  --role                 include role template
  --env                  include environment template
`)
		return nil
	}

	fmt.Println("Running init command...")
	// TODO: Implement init logic with scaffolding
	return fmt.Errorf("init not yet implemented")
}

func (c *CLI) runData(args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(`Usage: monad data [options]

Show contextual template data derived from basis.

Options:
  -h, --help             show this help message
  --json                 output as JSON
`)
		return nil
	}

	fmt.Println("Running data command...")
	// TODO: Implement data display logic
	return fmt.Errorf("data not yet implemented")
}

//
// Global option parsing
//

func (c *CLI) parseGlobalOptions(args []string) ([]string, error) {
	var remainingArgs []string
	
	for i := 0; i < len(args); i++ {
		arg := args[i]
		
		// Handle help flags
		if arg == "-h" || arg == "--help" {
			c.showHelp()
			return nil, nil
		}
		
		// Handle global options with values
		if strings.HasPrefix(arg, "--") {
			var optName, optValue string
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
				optName, optValue = parts[0], parts[1]
			} else {
				optName = arg
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
					optValue = args[i]
				} else {
					return nil, fmt.Errorf("option %s requires a value", optName)
				}
			}
			
			if err := c.setGlobalOption(optName, optValue); err != nil {
				return nil, err
			}
			continue
		}
		
		// Not a global option, add to remaining args
		remainingArgs = append(remainingArgs, arg)
	}
	
	return remainingArgs, nil
}

func (c *CLI) setGlobalOption(name, value string) error {
	// Handle CLI-only options
	if name == "--chdir" {
		c.chdir = value
		return nil
	}
	
	// Use annotation system to resolve flag to env var for basis options
	envVar, found := FlagToEnvVar(BasisConfig{}, name)
	if !found {
		return fmt.Errorf("unknown basis option: %s", name)
	}
	
	// Export to environment variable
	os.Setenv(envVar, value)
	return nil
}

//
// Helper functions
//

func getEnvOrDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}

func getBasename() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return filepath.Base(wd)
}