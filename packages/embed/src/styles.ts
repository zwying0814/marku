export const baseStyles = `
:host {
  --marku-font: system-ui, -apple-system, "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
  --marku-font-size: 15px;
  --marku-text: #1d1d20;
  --marku-muted: #6a6a74;
  --marku-bg: transparent;
  --marku-surface: #ffffff;
  --marku-border: rgba(20, 20, 24, 0.12);
  --marku-accent: #2f6f4e;
  --marku-accent-text: #ffffff;
  --marku-danger: #b42318;
  --marku-radius: 12px;
  --marku-gap: 16px;
  --marku-max-width: 720px;

  display: block;
  color: var(--marku-text);
  background: var(--marku-bg);
  font-family: var(--marku-font);
  font-size: var(--marku-font-size);
  line-height: 1.65;
  color-scheme: light dark;
  -webkit-font-smoothing: antialiased;
}

@media (prefers-color-scheme: dark) {
  :host {
    --marku-text: #e8e8ec;
    --marku-muted: #a0a0ab;
    --marku-surface: #17171b;
    --marku-border: rgba(255, 255, 255, 0.14);
    --marku-accent: #6cc59a;
    --marku-accent-text: #0d1a14;
    --marku-danger: #ff8a80;
  }
}

* {
  box-sizing: border-box;
}

.marku-container {
  max-width: var(--marku-max-width);
  margin: 0 auto;
}

.marku-error {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 12px;
  padding: 12px 14px;
  border: 1px solid var(--marku-border);
  border-left: 3px solid var(--marku-danger);
  border-radius: var(--marku-radius);
  color: var(--marku-danger);
  background: color-mix(in srgb, var(--marku-surface) 88%, var(--marku-danger));
}

button {
  font: inherit;
  cursor: pointer;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.marku-button {
  border: 1px solid var(--marku-border);
  border-radius: calc(var(--marku-radius) - 4px);
  background: var(--marku-surface);
  color: var(--marku-text);
  padding: 7px 14px;
  transition: border-color 120ms ease, background 120ms ease;
}

.marku-button:hover:not(:disabled) {
  border-color: var(--marku-accent);
}

.marku-button--primary {
  border-color: var(--marku-accent);
  background: var(--marku-accent);
  color: var(--marku-accent-text);
  font-weight: 600;
}

.marku-link-button {
  border: 0;
  background: none;
  color: var(--marku-muted);
  padding: 0;
  font-size: 0.85em;
}

.marku-link-button:hover {
  color: var(--marku-accent);
}

.marku-muted {
  color: var(--marku-muted);
}

.marku-field {
  display: grid;
  gap: 6px;
}

.marku-field label {
  font-size: 0.85em;
  color: var(--marku-muted);
}

.marku-field input,
.marku-field textarea {
  width: 100%;
  font: inherit;
  color: var(--marku-text);
  background: var(--marku-surface);
  border: 1px solid var(--marku-border);
  border-radius: calc(var(--marku-radius) - 4px);
  padding: 9px 12px;
}

.marku-field input:focus-visible,
.marku-field textarea:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--marku-accent) 60%, transparent);
  outline-offset: 1px;
  border-color: var(--marku-accent);
}

.marku-field textarea {
  min-height: 116px;
  resize: vertical;
}

.marku-honeypot {
  position: absolute;
  left: -10000px;
  width: 1px;
  height: 1px;
  overflow: hidden;
}

.marku-status {
  margin: 0;
  font-size: 0.88em;
  color: var(--marku-muted);
}

.marku-status[data-tone="success"] {
  color: var(--marku-accent);
}

.marku-status[data-tone="error"] {
  color: var(--marku-danger);
}

.marku-skeleton {
  display: grid;
  gap: 10px;
}

.marku-skeleton span {
  display: block;
  height: 12px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--marku-muted) 22%, transparent);
  animation: marku-pulse 1.4s ease-in-out infinite;
}

.marku-skeleton span:nth-child(2) {
  width: 82%;
}

.marku-skeleton span:nth-child(3) {
  width: 60%;
}

@keyframes marku-pulse {
  0%, 100% { opacity: 0.45; }
  50% { opacity: 0.9; }
}

@media (prefers-reduced-motion: reduce) {
  .marku-skeleton span {
    animation: none;
  }
}
`

export const commentsStyles = `
.marku-comments {
  display: grid;
  gap: calc(var(--marku-gap) * 1.25);
}

.marku-comments__header {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin: 0;
}

.marku-comments__heading {
  margin: 0;
  font-size: 1.05em;
  letter-spacing: 0.01em;
}

.marku-comments__count {
  color: var(--marku-muted);
  font-size: 0.85em;
}

.marku-comments__list {
  display: grid;
  gap: var(--marku-gap);
}

.marku-comment {
  display: grid;
  gap: 8px;
  padding: 14px 16px;
  border: 1px solid var(--marku-border);
  border-radius: var(--marku-radius);
  background: var(--marku-surface);
}

.marku-comment__head {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 8px;
}

.marku-comment__author {
  font-weight: 600;
}

.marku-comment__badge {
  font-size: 0.72em;
  padding: 1px 7px;
  border-radius: 999px;
  border: 1px solid var(--marku-accent);
  color: var(--marku-accent);
}

.marku-comment__time {
  margin-left: auto;
  color: var(--marku-muted);
  font-size: 0.8em;
}

.marku-comment__content {
  margin: 0;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.marku-comment__replies {
  display: grid;
  gap: 10px;
  margin-top: 4px;
  padding-left: 14px;
  border-left: 2px solid var(--marku-border);
}

.marku-comments__empty {
  margin: 0;
  padding: 22px 16px;
  text-align: center;
  color: var(--marku-muted);
  border: 1px dashed var(--marku-border);
  border-radius: var(--marku-radius);
}

.marku-comments__form {
  display: grid;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--marku-border);
  border-radius: var(--marku-radius);
  background: color-mix(in srgb, var(--marku-surface) 96%, var(--marku-accent));
}

.marku-comments__form-row {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.marku-comments__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.marku-comments__actions .marku-status {
  margin-left: auto;
}

.marku-comments__reply-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.88em;
  color: var(--marku-muted);
}

.marku-comments__more {
  justify-self: center;
}
`

export const likeStyles = `
.marku-like {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.marku-like__button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  border: 1px solid var(--marku-border);
  border-radius: 999px;
  padding: 8px 16px;
  background: var(--marku-surface);
  color: var(--marku-text);
  transition: border-color 120ms ease, color 120ms ease, transform 120ms ease;
}

.marku-like__button:hover:not(:disabled) {
  border-color: var(--marku-accent);
}

.marku-like__button[aria-pressed="true"] {
  border-color: var(--marku-accent);
  color: var(--marku-accent);
}

.marku-like__icon {
  font-size: 1.05em;
  line-height: 1;
}

.marku-like__count {
  font-variant-numeric: tabular-nums;
  color: var(--marku-muted);
}

.marku-like__button[aria-pressed="true"] .marku-like__count {
  color: inherit;
}
`

export const statsStyles = `
.marku-stats {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 6px 10px;
  color: var(--marku-muted);
  font-size: 0.9em;
}

.marku-stats__item {
  display: inline-flex;
  align-items: baseline;
  gap: 4px;
}

.marku-stats__value {
  color: var(--marku-text);
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}

.marku-stats__separator {
  opacity: 0.5;
}
`
