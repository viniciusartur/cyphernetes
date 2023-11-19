package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
	cobra "github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var ShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Launch an interactive shell",
	Run:   runShell,
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("MATCH"),
	readline.PcItem("match"),
	readline.PcItem("RETURN"),
	readline.PcItem("return"),
	readline.PcItem("help"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func shellPrompt() string {
	ns := Namespace
	color := "32"
	if ns == "" {
		ns = "ALL NAMESPACES"
		color = "31"
	}
	// Get the name of the current Kubernetes context
	context, err := getCurrentContext()
	if err != nil {
		fmt.Println("Error getting current context: ", err)
		return ""
	}

	return fmt.Sprintf("\033[%sm(%s) %s »\033[0m ", color, context, ns)
}

func getCurrentContext() (string, error) {
	// Use the local kubeconfig context
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: clientcmd.RecommendedHomeFile},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()
	if err != nil {
		fmt.Println("Error creating in-cluster config")
		return "", err
	}
	currentContextName := config.CurrentContext
	return currentContextName, nil
}

func runShell(cmd *cobra.Command, args []string) {
	historyFile := os.Getenv("HOME") + "/.cyphernetes_history"
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          shellPrompt(),
		HistoryFile:     historyFile,
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()
	rl.CaptureExitSignal()

	fmt.Println("Cyphernetes Interactive Shell")
	fmt.Println("Type 'exit' or press Ctrl-D to exit")
	fmt.Println("Type 'help' for information on how to use the shell")

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF, Ctrl-D
			break
		}

		input := strings.TrimSpace(line)
		if input == "exit" {
			break
		}

		// if input starts with '\n '
		if strings.HasPrefix(input, "\\n ") {
			input = strings.TrimPrefix(input, "\\n ")
			if strings.ToLower(input) == "all" {
				Namespace = ""
			} else {
				Namespace = strings.ToLower(input)
			}
			rl.SetPrompt(shellPrompt())
		} else if input == "help" {
			fmt.Println("Cyphernetes Interactive Shell")
			fmt.Println("exit           - Exit the shell")
			fmt.Println("help           - Print this help message")
			fmt.Println("\\n <namespace> - Change the namespace")
			fmt.Println("\\n all         - Query all namespaces")
		} else if input != "" {
			// Process the input if not empty
			result, err := processQuery(input)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
			} else {
				fmt.Println(result)
			}
		}
		// Add input to history
		rl.SaveHistory(input)
	}
}

// Execute the query against the Kubernetes API.
var executor = GetQueryExecutorInstance()

func processQuery(query string) (string, error) {
	// Parse the query to get an AST.
	ast, err := ParseQuery(query)
	if err != nil {
		// Handle error.
		fmt.Println("Error parsing query: ", err)
		return "", err
	}

	results, err := executor.Execute(ast)
	if err != nil {
		// Handle error.
		fmt.Println("Error executing query: ", err)
		return "", err
	}
	// Print the results as pretty JSON.
	json, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		// Handle error.
		fmt.Println("Error marshalling results: ", err)
		return "", err
	}
	return string(json), nil
}

func init() {
	rootCmd.AddCommand(ShellCmd)

	// Here you can define flags and configuration settings for the 'shell' subcommand if needed
}
