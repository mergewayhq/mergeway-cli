import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  Trace,
  TransportKind,
} from "vscode-languageclient/node";

let client: LanguageClient | undefined;
let outputChannel: vscode.OutputChannel | undefined;
let extensionContext: vscode.ExtensionContext | undefined;
let fileWatcher: vscode.FileSystemWatcher | undefined;

export async function activate(context: vscode.ExtensionContext): Promise<void> {
  extensionContext = context;
  outputChannel = vscode.window.createOutputChannel("Mergeway");
  context.subscriptions.push(outputChannel);

  appendLine("Activating Mergeway extension.");

  context.subscriptions.push(
    vscode.commands.registerCommand("mergeway.restartLanguageServer", async () => {
      appendLine("Manual restart requested.");
      await restartLanguageServer();
    }),
  );

  context.subscriptions.push(
    vscode.workspace.onDidChangeConfiguration(async (event) => {
      if (event.affectsConfiguration("mergeway.lsp.trace.server")) {
        applyTraceSetting();
      }

      if (event.affectsConfiguration("mergeway.lsp.path")) {
        appendLine("mergeway.lsp.path changed. Restarting language server.");
        await restartLanguageServer();
      }
    }),
  );

  const hasConfig = await workspaceContainsMergewayConfig();
  if (!hasConfig) {
    appendLine("No mergeway.yaml or mergeway.yml found. Language server will not start.");
    return;
  }

  await startLanguageServer();
}

export async function deactivate(): Promise<void> {
  await stopLanguageServer();
  appendLine("Deactivated Mergeway extension.");
}

async function restartLanguageServer(): Promise<void> {
  await stopLanguageServer();
  await startLanguageServer();
}

async function startLanguageServer(): Promise<void> {
  if (client) {
    appendLine("Language server is already running.");
    return;
  }

  if (!(await workspaceContainsMergewayConfig())) {
    appendLine("No Mergeway config found during startup. Skipping language server start.");
    return;
  }

  const lspPath = await validateConfiguredLspPath();
  if (!lspPath) {
    return;
  }

  fileWatcher = vscode.workspace.createFileSystemWatcher("**/*.{yaml,yml,json}");
  extensionContext?.subscriptions.push(fileWatcher);
  const channel = requireOutputChannel();

  const serverOptions: ServerOptions = {
    run: {
      command: lspPath,
      transport: TransportKind.stdio,
    },
    debug: {
      command: lspPath,
      transport: TransportKind.stdio,
      options: {
        env: {
          ...process.env,
          MERGEWAY_LSP_LOG_STDERR: "1",
          MERGEWAY_LSP_LOG_LEVEL: "debug",
        },
      },
    },
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [
      { scheme: "file", language: "yaml" },
      { scheme: "file", language: "json" },
    ],
    synchronize: {
      fileEvents: fileWatcher,
    },
    outputChannel: channel,
    traceOutputChannel: channel,
    initializationOptions: {
      configFiles: ["mergeway.yaml", "mergeway.yml"],
    },
  };

  client = new LanguageClient(
    "mergewayLanguageServer",
    "Mergeway Language Server",
    serverOptions,
    clientOptions,
  );

  if (extensionContext) {
    extensionContext.subscriptions.push(client);
  }

  appendLine(`Starting Mergeway LSP: ${lspPath}`);
  await client.start();
  applyTraceSetting();
  appendLine("Mergeway LSP started.");
}

async function stopLanguageServer(): Promise<void> {
  if (!client) {
    return;
  }

  appendLine("Stopping Mergeway LSP.");
  const currentClient = client;
  client = undefined;
  await currentClient.stop();
  appendLine("Mergeway LSP stopped.");

  if (fileWatcher) {
    fileWatcher.dispose();
    fileWatcher = undefined;
  }
}

async function workspaceContainsMergewayConfig(): Promise<boolean> {
  const files = await vscode.workspace.findFiles(
    "{mergeway.yaml,mergeway.yml}",
    "**/{node_modules,.git}/**",
    1,
  );

  return files.length > 0;
}

function getConfiguredLspPath(): string | undefined {
  const value = vscode.workspace.getConfiguration("mergeway").get<string>("lsp.path");
  if (!value) {
    return undefined;
  }

  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : undefined;
}

async function validateConfiguredLspPath(): Promise<string | undefined> {
  const lspPath = getConfiguredLspPath();
  if (!lspPath) {
    await showMissingPathMessage();
    return undefined;
  }

  if (!path.isAbsolute(lspPath)) {
    const message = `mergeway.lsp.path must be an absolute path. Current value: ${lspPath}`;
    appendLine(message);
    await vscode.window.showErrorMessage(message);
    return undefined;
  }

  let stat: fs.Stats;
  try {
    stat = await fs.promises.stat(lspPath);
  } catch {
    const message = `Mergeway LSP binary not found at configured path: ${lspPath}`;
    appendLine(message);
    await vscode.window.showErrorMessage(message);
    return undefined;
  }

  if (!stat.isFile()) {
    const message = `mergeway.lsp.path must point to a file: ${lspPath}`;
    appendLine(message);
    await vscode.window.showErrorMessage(message);
    return undefined;
  }

  if (process.platform !== "win32") {
    try {
      await fs.promises.access(lspPath, fs.constants.X_OK);
    } catch {
      const message = `mergeway.lsp.path is not executable: ${lspPath}`;
      appendLine(message);
      await vscode.window.showErrorMessage(message);
      return undefined;
    }
  }

  return lspPath;
}

async function showMissingPathMessage(): Promise<void> {
  const selection = await vscode.window.showWarningMessage(
    "Mergeway LSP path is not configured. Set mergeway.lsp.path to the absolute path of the mergeway-lsp binary.",
    "Open Settings",
  );

  if (selection === "Open Settings") {
    await vscode.commands.executeCommand(
      "workbench.action.openSettings",
      "mergeway.lsp.path",
    );
  }
}

function applyTraceSetting(): void {
  if (!client) {
    return;
  }

  const configuredTrace = vscode.workspace
    .getConfiguration("mergeway")
    .get<string>("lsp.trace.server", "off");

  const trace = toTrace(configuredTrace);
  client.setTrace(trace);
  appendLine(`Set Mergeway LSP trace to ${configuredTrace}.`);
}

function toTrace(value: string): Trace {
  switch (value) {
    case "messages":
      return Trace.Messages;
    case "verbose":
      return Trace.Verbose;
    default:
      return Trace.Off;
  }
}

function appendLine(message: string): void {
  outputChannel?.appendLine(message);
}

function requireOutputChannel(): vscode.OutputChannel {
  if (!outputChannel) {
    throw new Error("Mergeway output channel is not initialized.");
  }

  return outputChannel;
}
