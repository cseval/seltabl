package cmds

import (
	"context"
	"io"

	"github.com/spf13/cobra"
)

// NewGenerateCmd returns the generate command
func NewGenerateCmd(_ context.Context, _ io.Writer, _ io.Reader) *cobra.Command {
	var url string
	cmd := &cobra.Command{
		Use:   "generate", // the name of the command
		Short: "Generates a new seltabl struct for a given url.",
		Long: `
Generates a new seltabl struct for a given url.

The command will create a new package in the current directory with the name "seltabl".
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SetOutput(w)
			cmd.SetIn(r)
			cmd.SetErr(w)
			cmd.SetContext(ctx)
			llmAPIKey := os.Getenv("LLM_API_KEY")
			if llmAPIKey == "" {
				return fmt.Errorf("LLM_API_KEY is not set")
			}
			client := llm.CreateClient(
				url,
				llmAPIKey,
			)
			state, err := analysis.NewState()
			if err != nil {
				return fmt.Errorf("failed to create state: %w", err)
			}
			sels, err := analysis.GetSelectors(
				ctx,
				&state,
				url,
				ignores,
			)
			if err != nil {
				return fmt.Errorf("failed to get selectors: %w", err)
			}
			body, err := getURL(url)
			if err != nil {
				return fmt.Errorf("failed to get url: %w", err)
			}
			basePrompt, err := prompts.NewBasePrompt(
				sels,
				string(body),
				url,
			)
			history := []openai.ChatCompletionMessage{{
				Role:    openai.ChatMessageRoleUser,
				Content: basePrompt,
			}}
			content, history, err := Chat(
				ctx,
				client,
				llmModel,
				history,
				basePrompt,
			)
			if err != nil {
				return fmt.Errorf("failed to create chat completion: %w", err)
			}
			err = verify(ctx, cmd, name)
			if err == nil {
				print("verified generated struct")
				return nil
			}
			_, err = prompts.NewErrPrompt(string(body), content, url, err)
			if err != nil {
				return fmt.Errorf("failed to create err prompt: %w", err)
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&url, "url", "u", "", "The url for which to generate a seltabl struct.")
	cmd.PersistentFlags().StringVarP(&url, "name", "n", "", "The name of the struct to generate.")
	registerCompletionFuncForGlobalFlags(cmd)
	return cmd
}

// registerCompletionFuncForGlobalFlags registers a completion function for the global flags
func registerCompletionFuncForGlobalFlags(cmd *cobra.Command) (err error) {
	err = cmd.RegisterFlagCompletionFunc(
		"url",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"https://github.com/conneroisu/seltabl/blob/main/testdata/ab_num_table.html"}, cobra.ShellCompDirectiveDefault
		},
	)
	cmd.PersistentFlags().StringVarP(
		&llmModel,
		"llm-model",
		"m",
		"llama3-70b-8192",
		"The name of the llm model to use for generating the struct.",
	)
	cmd.PersistentFlags().StringSliceVarP(
		&ignores,
		"ignore",
		"i",
		[]string{"script", "style", "link", "img", "footer", "header"},
		"The elements to ignore when generating the struct.",
	)
	return cmd, nil
}

// verify verifies the generated struct
func verify(
	ctx context.Context,
	cmd *cobra.Command,
	name string,
) error {
	err := NewVetCmd(ctx, os.Stdout).RunE(cmd, []string{name + ".go"})
	if err == nil {
		fmt.Printf("Generated %s\n", name)
		return nil
	}
	return fmt.Errorf("failed to vet generated struct: %w", err)
}

// writeFile writes a file to the given path
func writeFile(name string, content string) error {
	f, err := os.Create(name + ".go")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// Chat is a struct for a chat
func Chat(
	ctx context.Context,
	client *openai.Client,
	model string,
	history []openai.ChatCompletionMessage,
	prompt string,
) (string, []openai.ChatCompletionMessage, error) {
	completion, err := client.CreateChatCompletion(
		ctx, openai.ChatCompletionRequest{
			Model: model,
			Messages: append(history, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			}),
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: "json",
			},
		})
	if err != nil {
		return "", history, fmt.Errorf("failed to create chat completion: %w", err)
	}
	content := completion.Choices[0].Message.Content
	history = append(history, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: content})
	return content, history, nil
}

// getURL gets the url and returns the body
func getURL(url string) ([]byte, error) {
	cli := http.DefaultClient
	resp, err := cli.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get url: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get url: %s", resp.Status)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	return body, nil
}
