package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/conneroisu/seltabl/tools/seltabls/pkg/analysis"
	"github.com/conneroisu/seltabl/tools/seltabls/pkg/lsp"
	"github.com/conneroisu/seltabl/tools/seltabls/pkg/lsp/methods"
	"github.com/conneroisu/seltabl/tools/seltabls/pkg/rpc"
)

// HandleMessage handles a message sent from the client to the language server.
// It parses the message and returns with a response.
func HandleMessage(
	ctx context.Context,
	state *analysis.State,
	msg rpc.BaseMessage,
) (response rpc.MethodActor, err error) {
	hCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	for {
		select {
		case <-hCtx.Done():
			return nil, fmt.Errorf("context cancelled: %w", hCtx.Err())
		default:
			switch methods.Method(msg.Method) {
			case methods.MethodInitialize:
				var request lsp.InitializeRequest
				err = json.Unmarshal([]byte(msg.Content), &request)
				if err != nil {
					return nil, fmt.Errorf(
						"decode initialize request (initialize) failed: %w",
						err,
					)
				}
				return lsp.NewInitializeResponse(&request)
			case methods.MethodRequestTextDocumentDidOpen:
				var request lsp.NotificationDidOpenTextDocument
				err = json.Unmarshal(msg.Content, &request)
				if err != nil {
					return nil, fmt.Errorf(
						"decode (textDocument/didOpen) request failed: %w",
						err,
					)
				}
				return analysis.OpenDocument(hCtx, state, request)
			case methods.MethodRequestTextDocumentCompletion:
				var request lsp.CompletionRequest
				err = json.Unmarshal(msg.Content, &request)
				if err != nil {
					return nil, fmt.Errorf(
						"failed to unmarshal completion request (textDocument/completion): %w",
						err,
					)
				}
				return analysis.CreateTextDocumentCompletion(
					hCtx,
					state,
					request,
				)
			case methods.MethodRequestTextDocumentHover:
				var request lsp.HoverRequest
				err = json.Unmarshal(msg.Content, &request)
				if err != nil {
					return nil, fmt.Errorf(
						"failed unmarshal of hover request (): %w",
						err,
					)
				}
				response, err = analysis.NewHoverResponse(request, state)
				if err != nil {
					return nil, fmt.Errorf("failed to get hover: %w", err)
				}
				return response, nil
			case methods.MethodRequestTextDocumentCodeAction:
				var request lsp.CodeActionRequest
				err = json.Unmarshal(msg.Content, &request)
				if err != nil {
					return nil, fmt.Errorf(
						"failed to unmarshal of codeAction request (textDocument/codeAction): %w",
						err,
					)
				}
				return analysis.TextDocumentCodeAction(
					hCtx,
					request,
					state,
				)
			case methods.MethodShutdown:
				var request lsp.ShutdownRequest
				err = json.Unmarshal([]byte(msg.Content), &request)
				if err != nil {
					return nil, fmt.Errorf(
						"decode (shutdown) request failed: %w",
						err,
					)
				}
				return lsp.NewShutdownResponse(request, nil)
			case methods.MethodCancelRequest:
				var request lsp.CancelRequest
				err = json.Unmarshal(msg.Content, &request)
				if err != nil {
					return nil, fmt.Errorf(
						"failed to unmarshal cancel request ($/cancelRequest): %w",
						err,
					)
				}
				response, err = state.CancelRequest(request)
				if err != nil || response == nil {
					return nil, fmt.Errorf("failed to cancel request: %w", err)
				}
				return response, nil

			case methods.MethodNotificationInitialized:
				var request lsp.InitializedParamsRequest
				err = json.Unmarshal([]byte(msg.Content), &request)
				if err != nil {
					return nil, fmt.Errorf(
						"decode (initialized) request failed: %w",
						err,
					)
				}
				return nil, nil
			case methods.MethodNotificationExit:
				os.Exit(0)
				return nil, nil
			case methods.MethodNotificationTextDocumentDidSave:
				var request lsp.DidSaveTextDocumentParamsNotification
				err = json.Unmarshal([]byte(msg.Content), &request)
				if err != nil {
					return nil, fmt.Errorf(
						"decode (didSave) request failed: %w",
						err,
					)
				}
				u, err := url.Parse(request.Params.TextDocument.URI)
				if err != nil {
					return nil, fmt.Errorf("failed to parse uri: %w", err)
				}
				read, err := os.ReadFile(u.Path)
				if err != nil {
					return nil, fmt.Errorf("failed to read file: %w", err)
				}
				state.Documents[request.Params.TextDocument.URI] = string(read)
				return nil, nil
			case methods.NotificationTextDocumentDidClose:
				var request lsp.DidCloseTextDocumentParamsNotification
				if err = json.Unmarshal([]byte(msg.Content), &request); err != nil {
					return nil, fmt.Errorf(
						"decode (didClose) request failed: %w",
						err,
					)
				}
			case methods.NotificationMethodTextDocumentDidChange:
				var request lsp.TextDocumentDidChangeNotification
				err = json.Unmarshal(msg.Content, &request)
				if err != nil {
					return nil, fmt.Errorf(
						"decode (textDocument/didChange) request failed: %w",
						err,
					)
				}
				return analysis.UpdateDocument(hCtx, state, &request)
			default:
				return nil, fmt.Errorf("unknown method: %s", msg.Method)
			}
		}
	}
}
