package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/conneroisu/seltabl/tools/seltabls/pkg/lsp"
	"github.com/conneroisu/seltabl/tools/seltabls/pkg/rpc"
)

// HandleMessage handles a message sent from the client to the language server.
// It parses the message and returns with a response.
func (s *Root) HandleMessage(
	ctx context.Context,
	msg []byte,
) error {
	var response interface{}
	var err error
	method, contents, err := rpc.DecodeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}
	switch method {
	case "initialize":
		var request lsp.InitializeRequest
		if err = json.Unmarshal([]byte(contents), &request); err != nil {
			return fmt.Errorf(
				"decode initialize request (initialize) failed: %w",
				err,
			)
		}
		response = lsp.NewInitializeResponse(request.ID)
		err = s.writeResponse(ctx, method, response)
		if err != nil {
			return fmt.Errorf("failed to write (initialize) response: %w", err)
		}
	case "initialized":
		var request lsp.InitializedParamsRequest
		if err = json.Unmarshal([]byte(contents), &request); err != nil {
			s.Logger.Fatal(
				"decode initialized request (initialized) failed",
				err,
			)
			return fmt.Errorf("decode (initialized) request failed: %w", err)
		}
	case "textDocument/didOpen":
		var request lsp.DidOpenTextDocumentNotification
		if err = json.Unmarshal(contents, &request); err != nil {
			return fmt.Errorf(
				"decode (textDocument/didOpen) request failed: %w",
				err,
			)
		}
		diagnostics, err := s.State.OpenDocument(
			request.Params.TextDocument.URI,
			&request.Params.TextDocument.Text,
		)
		if err != nil {
			return fmt.Errorf("failed to open document: %w", err)
		}
		response = lsp.PublishDiagnosticsNotification{
			Notification: lsp.Notification{
				RPC:    "2.0",
				Method: "textDocument/publishDiagnostics",
			},
			Params: lsp.PublishDiagnosticsParams{
				URI:         request.Params.TextDocument.URI,
				Diagnostics: diagnostics,
			},
		}
		err = s.writeResponse(ctx, method, response)
		if err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	case "textDocument/didClose":
		// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_didClose
		var request lsp.DidCloseTextDocumentParamsNotification
		if err = json.Unmarshal([]byte(contents), &request); err != nil {
			return fmt.Errorf("decode (didClose) request failed: %w", err)
		}
		s.State.Documents[request.Params.TextDocument.URI] = ""
	case "textDocument/completion":
		var request lsp.CompletionRequest
		err = json.Unmarshal(contents, &request)
		if err != nil {
			return fmt.Errorf(
				"failed to unmarshal completion request (textDocument/completion): %w",
				err,
			)
		}
		response, err = s.State.CreateTextDocumentCompletion(
			request.ID,
			request.Params.TextDocument,
			request.Params.Position,
		)
		if err != nil {
			return fmt.Errorf("failed to get completions: %w", err)
		}
		err = s.writeResponse(ctx, method, response)
		if err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	case "textDocument/didChange":
		var request lsp.TextDocumentDidChangeNotification
		err = json.Unmarshal(contents, &request)
		if err != nil {
			return fmt.Errorf(
				"decode (textDocument/didChange) request failed: %w",
				err,
			)
		}
		diagnostics := []lsp.Diagnostic{}
		for _, change := range request.Params.ContentChanges {
			diags, err := s.State.UpdateDocument(
				request.Params.TextDocument.URI,
				change.Text,
			)
			if err != nil {
				return fmt.Errorf("failed to update document: %w", err)
			}
			diagnostics = append(diagnostics, diags...)
		}
		response = lsp.PublishDiagnosticsNotification{
			Notification: lsp.Notification{
				RPC:    "2.0",
				Method: "textDocument/publishDiagnostics",
			},
			Params: lsp.PublishDiagnosticsParams{
				URI:         request.Params.TextDocument.URI,
				Diagnostics: diagnostics,
			},
		}
		if err = s.writeResponse(ctx, method, response); err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	case "textDocument/hover":
		var request lsp.HoverRequest
		err = json.Unmarshal(contents, &request)
		if err != nil {
			return fmt.Errorf("failed unmarshal of hover request (): %w", err)
		}
		response, err = s.State.Hover(
			request.ID,
			request.Params.TextDocument.URI,
			request.Params.Position,
		)
		if err != nil {
			return fmt.Errorf("failed to get hover: %w", err)
		}
		err = s.writeResponse(ctx, method, response)
		if err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	case "textDocument/codeAction":
		var request lsp.CodeActionRequest
		err = json.Unmarshal(contents, &request)
		if err != nil {
			return fmt.Errorf(
				"failed to unmarshal of codeAction request (textDocument/codeAction): %w",
				err,
			)
		}
		response = s.State.TextDocumentCodeAction(
			request.ID,
			request.Params.TextDocument.URI,
		)
		err = s.writeResponse(ctx, method, response)
		if err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	case "textDocument/didSave":
		// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_didSave
	case "shutdown":
		var request lsp.ShutdownRequest
		if err = json.Unmarshal([]byte(contents), &request); err != nil {
			return fmt.Errorf("decode (shutdown) request failed: %w", err)
		}
		response = lsp.ShutdownResponse{
			Response: lsp.Response{
				RPC: "2.0",
				ID:  request.ID,
			},
		}
		err = s.writeResponse(ctx, method, response)
		if err != nil {
			return fmt.Errorf("write (shutdown) response failed: %w", err)
		}
		os.Exit(0)
	case "$/cancelRequest":
		var request lsp.CancelRequest
		err = json.Unmarshal(contents, &request)
		if err != nil {
			return fmt.Errorf(
				"failed to unmarshal cancel request ($/cancelRequest): %w",
				err,
			)
		}
		response, err = s.State.CancelRequest(request.ID)
		if err != nil {
			return fmt.Errorf("failed to cancel request: %w", err)
		}
		err = s.writeResponse(ctx, method, response)
		if err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	case "exit":
		os.Exit(0)
		return nil
	default:
		return fmt.Errorf("unknown method: %s", method)
	}
	if response != nil {
		enc, err := rpc.EncodeMessage(response)
		if err != nil {
			return fmt.Errorf("failed to encode message: %w", err)
		}
		s.Logger.Printf(
			"Received message (%s) err: [%s] response: `%s` contents: %s",
			method,
			err,
			strings.Replace(enc, "\n", " ", -1),
			contents,
		)
		s.State.Logger.Printf(
			"Received message (%s) err: [%s] response: `%s` contents: %s",
			method,
			err,
			strings.Replace(enc, "\n", " ", -1),
			contents,
		)
		return nil
	}
	s.Logger.Printf("no response for %s", method)
	return nil
}