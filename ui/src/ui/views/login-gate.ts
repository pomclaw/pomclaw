import { html } from "lit";
import { t } from "../../i18n/index.ts";
import type { AppViewState } from "../app-view-state.ts";
import { icons } from "../icons.ts";
import { normalizeBasePath } from "../navigation.ts";
import { agentLogoUrl } from "./agents-utils.ts";

export function renderLoginGate(state: AppViewState) {
  const basePath = normalizeBasePath(state.basePath ?? "");
  const faviconSrc = agentLogoUrl(basePath);

  return html`
    <div class="login-gate">
      <div class="login-gate__card">
        <div class="login-gate__header">
          <img class="login-gate__logo" src=${faviconSrc} alt="PomClaw" />
          <div class="login-gate__title">PomClaw</div>
          <div class="login-gate__sub">请输入用户账号登录</div>
        </div>
        <div class="login-gate__form">
          <label class="field">
            <span>用户账号</span>
            <input
              .value=${state.settings.token || ""}
              @input=${(e: Event) => {
                const v = (e.target as HTMLInputElement).value;
                state.applySettings({ ...state.settings, token: v });
              }}
              placeholder="请输入用户账号"
              @keydown=${(e: KeyboardEvent) => {
                if (e.key === "Enter") {
                  state.connect();
                }
              }}
            />
          </label>
          <button
            class="btn primary login-gate__connect"
            @click=${() => state.connect()}
          >
            登录
          </button>
        </div>
        ${
          state.lastError
            ? html`<div class="callout danger" style="margin-top: 14px;">
                <div>${state.lastError}</div>
              </div>`
            : ""
        }
      </div>
    </div>
  `;
}
