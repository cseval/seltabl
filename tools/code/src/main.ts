import * as vscode from "vscode";
import {
  CancellationToken,
  CloseAction,
  CloseHandlerResult,
  CompletionItemKind,
  ConfigurationParams,
  ConfigurationRequest,
  ErrorAction,
  ErrorHandlerResult,
  Message,
  ProvideCompletionItemsSignature,
  ProvideDocumentFormattingEditsSignature,
  ResponseError,
} from "vscode-languageclient";
import fs from "fs/promises";
import path from "path";
import { LanguageClient } from "vscode-languageclient/node";
import { lookpath } from "lookpath";
import { CustomLanguageClient } from "./custom-client";

export async function activate(ctx: vscode.ExtensionContext) {
  try {
    ctx.subscriptions.push(
      vscode.commands.registerCommand(
        "seltabls.restartServer",
        startLanguageClient
      )
    );

    await startLanguageClient();
  } catch (err) {
    const msg = err && (err as Error) ? (err as Error).message : "unknown";
    vscode.window.showErrorMessage(`error initializing seltabls LSP: ${msg}`);
  }
}

interface seltablsCtx {
  languageClient?: LanguageClient;
}

const ctx: seltablsCtx = {};

const seltablsLocations = [
  path.join(process.env.GOBIN ?? "", "seltabls"),
  path.join(process.env.GOBIN ?? "", "seltabls.exe"),
  path.join(process.env.GOPATH ?? "", "bin", "seltabls"),
  path.join(process.env.GOPATH ?? "", "bin", "seltabls.exe"),
  path.join(process.env.GOROOT || "", "bin", "seltabls"),
  path.join(process.env.GOROOT || "", "bin", "seltabls.exe"),
  path.join(process.env.HOME || "", "bin", "seltabls"),
  path.join(process.env.HOME || "", "bin", "seltabls.exe"),
  path.join(process.env.HOME || "", "go", "bin", "seltabls"),
  path.join(process.env.HOME || "", "go", "bin", "seltabls.exe"),
  "/usr/local/bin/seltabls",
  "/usr/bin/seltabls",
  "/usr/local/go/bin/seltabls",
  "/usr/local/share/go/bin/seltabls",
  "/usr/share/go/bin/seltabls",
];

async function findSeltabls(): Promise<string> {
  const linuxName = await lookpath("seltabls");
  if (linuxName) {
    return linuxName;
  }
  const windowsName = await lookpath("seltabls.exe");
  if (windowsName) {
    return windowsName;
  }
  for (const exe of seltablsLocations) {
    try {
      await fs.stat(exe);
      return exe;
    } catch (err) {
      // ignore
    }
  }
  throw new Error(
    `Could not find seltabls executable in path or in ${seltablsLocations.join(", ")}`
  );
}

async function stopLanguageClient() {
  const c = ctx.languageClient;
  ctx.languageClient = undefined;
  if (!c) return false;

  if (c.diagnostics) {
    c.diagnostics.clear();
  }
  // LanguageClient.stop may hang if the language server
  // crashes during shutdown before responding to the
  // shutdown request. Enforce client-side timeout.
  try {
    c.stop(2000);
  } catch (e) {
    c.outputChannel?.appendLine(`Failed to stop client: ${e}`);
  }
}

async function startLanguageClient() {
  ctx.languageClient = await buildLanguageClient();
  await ctx.languageClient.start();
}

export async function buildLanguageClient(): Promise<LanguageClient> {
  const documentSelector = [{ language: "go", scheme: "file" }];

  const args: Array<string> = ["lsp"];
  const seltablsPath = await findSeltabls();

  if (ctx.languageClient) {
    await stopLanguageClient();
  }

  vscode.window.setStatusBarMessage(
    `Starting LSP: seltabls ${args.join(" ")}`,
    3000
  );

  const c = new CustomLanguageClient(
    "seltabls", // id
    "seltabls",
    {
      command: seltablsPath,
      args,
    },
    {
      documentSelector,
      uriConverters: {
        // Apply file:/// scheme to all file paths.
        code2Protocol: (uri: vscode.Uri): string =>
          (uri.scheme ? uri : uri.with({ scheme: "file" })).toString(),
        protocol2Code: (uri: string) => vscode.Uri.parse(uri),
      },
      errorHandler: {
        error: (
          error: Error,
          message: Message,
          count: number
        ): ErrorHandlerResult => {
          // Allow 5 crashes before shutdown.
          if (count < 5) {
            return { action: ErrorAction.Continue };
          }
          vscode.window.showErrorMessage(
            `Error communicating with the language server: ${error}: ${message}.`
          );
          return { action: ErrorAction.Shutdown };
        },
        closed: (): CloseHandlerResult => ({
          action: CloseAction.DoNotRestart,
        }),
      },
      middleware: {
        provideDocumentFormattingEdits: async (
          document: vscode.TextDocument,
          options: vscode.FormattingOptions,
          token: vscode.CancellationToken,
          next: ProvideDocumentFormattingEditsSignature
        ) => {
          return next(document, options, token);
        },
        provideCompletionItem: async (
          document: vscode.TextDocument,
          position: vscode.Position,
          context: vscode.CompletionContext,
          token: vscode.CancellationToken,
          next: ProvideCompletionItemsSignature
        ) => {
          const list = await next(document, position, context, token);
          if (!list) {
            return list;
          }
          const items = Array.isArray(list) ? list : list.items;

          // Give all the candidates the same filterText to trick VSCode
          // into not reordering our candidates. All the candidates will
          // appear to be equally good matches, so VSCode's fuzzy
          // matching/ranking just maintains the natural "sortText"
          // ordering. We can only do this in tandem with
          // "incompleteResults" since otherwise client side filtering is
          // important.
          if (
            !Array.isArray(list) &&
            list.isIncomplete &&
            list.items.length > 1
          ) {
            let hardcodedFilterText = items[0].filterText;
            if (!hardcodedFilterText) {
              // tslint:disable:max-line-length
              // According to LSP spec,
              // https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_completion
              // if filterText is falsy, the `label` should be used.
              // But we observed that's not the case.
              // Even if vscode picked the label value, that would
              // cause to reorder candidates, which is not ideal.
              // Force to use non-empty `label`.
              // https://github.com/golang/vscode-go/issues/441
              hardcodedFilterText = items[0].label.toString();
            }
            for (const item of items) {
              item.filterText = hardcodedFilterText;
            }
          }
          // TODO(hyangah): when v1.42+ api is available, we can simplify
          // language-specific configuration lookup using the new
          // ConfigurationScope.
          //    const paramHintsEnabled = vscode.workspace.getConfiguration(
          //          'editor.parameterHints',
          //          { languageId: 'go', uri: document.uri });
          const editorParamHintsEnabled = vscode.workspace.getConfiguration(
            "editor.parameterHints",
            document.uri
          )["enabled"];
          const goParamHintsEnabled = vscode.workspace.getConfiguration(
            "[go]",
            document.uri
          )["editor.parameterHints.enabled"];
          let paramHintsEnabled = false;
          if (typeof goParamHintsEnabled === "undefined") {
            paramHintsEnabled = editorParamHintsEnabled;
          } else {
            paramHintsEnabled = goParamHintsEnabled;
          }
          // If the user has parameterHints (signature help) enabled,
          // trigger it for function or method completion items.
          if (paramHintsEnabled) {
            for (const item of items) {
              if (
                item.kind === CompletionItemKind.Method ||
                item.kind === CompletionItemKind.Function
              ) {
                item.command = {
                  title: "triggerParameterHints",
                  command: "editor.action.triggerParameterHints",
                };
              }
            }
          }
          return list;
        },
        // Keep track of the last file change in order to not prompt
        // user if they are actively working.
        didOpen: async (e, next) => next(e),
        didChange: async (e, next) => next(e),
        didClose: (e, next) => next(e),
        didSave: (e, next) => next(e),
        workspace: {
          configuration: async (
            params: ConfigurationParams,
            token: CancellationToken,
            next: ConfigurationRequest.HandlerSignature
          ): Promise<any[] | ResponseError<void>> => {
            const configs = await next(params, token);
            if (!configs || !Array.isArray(configs)) {
              return configs;
            }
            const ret = [] as any[];
            for (let i = 0; i < configs.length; i++) {
              let workspaceConfig = configs[i];
              console.log(workspaceConfig);
              ret.push(workspaceConfig);
            }
            return ret;
          },
        },
      },
    },
    false
  );
  return c;
}