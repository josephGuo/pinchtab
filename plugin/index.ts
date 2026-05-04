/**
 * Pinchtab OpenClaw Plugin
 *
 * Two tools:
 * - `pinchtab`: Full-featured browser control with all actions
 * - `browser`: OpenClaw-compatible simplified interface
 */

import type { PluginApi, PluginConfig, PluginTool } from "./types.js";
import { pinchtabToolSchema, pinchtabToolDescription, executePinchtabAction } from "./tools/pinchtab.js";
import { browserToolSchema, browserToolDescription, executeBrowserAction } from "./tools/browser.js";

function getConfig(api: PluginApi): PluginConfig {
  return (api.pluginConfig ?? api.config?.plugins?.entries?.pinchtab?.config ?? {}) as PluginConfig;
}

export default function register(api: PluginApi) {
  const cfg = getConfig(api);

  // Register the full-featured pinchtab tool
  const pinchtabTool = {
    name: "pinchtab",
    label: "PinchTab",
    description: pinchtabToolDescription,
    parameters: pinchtabToolSchema,
    async execute(_id: string, params: any) {
      return executePinchtabAction(getConfig(api), params);
    },
  } satisfies PluginTool;
  api.registerTool(pinchtabTool, { optional: true });

  // Register OpenClaw-compatible browser tool
  if (cfg.registerBrowserTool !== false) {
    const browserTool = {
      name: "browser",
      label: "Browser",
      description: browserToolDescription,
      parameters: browserToolSchema,
      async execute(_id: string, params: any) {
        return executeBrowserAction(getConfig(api), params);
      },
    } satisfies PluginTool;
    api.registerTool(browserTool, { optional: true });
  }
}
