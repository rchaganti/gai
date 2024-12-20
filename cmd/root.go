package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/rchaganti/gai/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	apiKey         string
	model          string
	promptFromFile string
	prompt         string
)
var rootCmd = &cobra.Command{
	Use:   "gai",
	Short: "A console UI for interacting with Google's Gemini AI",
	Long: `GAI is a console UI for interacting with Google's Gemini AI.
		You can converse with Gemini AI and get responses.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api-key")
		if apiKey == "" {
			slog.Error("API Key is required. Use --api-key or set GAI_API_KEY environment variable.")
			os.Exit(1)
		}

		model := viper.GetString("model")
		if !viper.IsSet("prompt-from-file") {
			if len(args) != 0 {
				prompt = args[0]
			} else {
				slog.Error("Prompt is required. Use --prompt-from-file or pass it as an argument.")
				os.Exit(1)
			}
		} else {
			filePath := viper.GetString("prompt-from-file")
			c, err := os.ReadFile(filePath)
			if err != nil {
				panic(err)
			}
			prompt = string(c)
		}

		tui.InvokeResponseTUI(apiKey, model, prompt)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "API Key for Gemini AI")
	viper.BindPFlag("api-key", rootCmd.Flags().Lookup("api-key"))
	viper.BindEnv("api-key", "GAI_API_KEY")

	rootCmd.Flags().StringVarP(&model, "model", "m", "gemini-pro", "Model to use for Gemini AI")
	viper.BindPFlag("model", rootCmd.Flags().Lookup("model"))
	viper.BindEnv("model", "GAI_MODEL")

	rootCmd.Flags().StringVarP(&promptFromFile, "prompt-from-file", "f", "", "Read prompt from file")
	viper.BindPFlag("prompt-from-file", rootCmd.Flags().Lookup("prompt-from-file"))
}
