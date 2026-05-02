import {
  useState,
  useRef,
  useCallback,
  type KeyboardEvent,
} from 'react'

interface InputBarProps {
  onSend: (text: string) => void
  onStop?: () => void
  disabled?: boolean
  isRunning?: boolean
  placeholder?: string
}

export function InputBar({
  onSend,
  onStop,
  disabled,
  isRunning,
  placeholder,
}: InputBarProps) {
  const [text, setText] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const handleSend = useCallback(() => {
    const hasContent = text.trim().length > 0
    if (!hasContent || disabled) return
    onSend(text.trim())
    setText('')
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }, [text, disabled, onSend])

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleInput = () => {
    const el = textareaRef.current
    if (!el) return
    el.style.height = 'auto'
    el.style.height = Math.min(el.scrollHeight, 160) + 'px'
  }

  const hasContent = text.trim().length > 0

  return (
    <div className="px-4 pb-4 pt-1 shrink-0">
      <div className="max-w-3xl mx-auto">
        <div
          className={[
            'flex items-end gap-0 bg-secondary rounded-2xl border transition-colors',
            'border-border focus-within:border-primary/40',
          ].join(' ')}
        >
          <textarea
            ref={textareaRef}
            value={text}
            onChange={(e) => {
              setText(e.target.value)
              handleInput()
            }}
            onKeyDown={handleKeyDown}
            placeholder={placeholder ?? 'Send a message...'}
            disabled={disabled}
            rows={1}
            className="flex-1 bg-transparent text-foreground text-base md:text-sm py-3 px-4 focus:outline-none placeholder:text-muted-foreground resize-none overflow-y-auto"
            style={{ maxHeight: 160 }}
          />

          <div className="p-2 shrink-0">
            {isRunning ? (
              <button
                onClick={onStop}
                className="w-10 h-10 flex items-center justify-center hover:opacity-90 transition-opacity"
                title="Stop generation"
              >
                <svg
                  className="absolute w-8 h-8 animate-spin"
                  viewBox="0 0 32 32"
                  fill="none"
                  style={{ animationDuration: '1.5s' }}
                >
                  <circle
                    cx="16"
                    cy="16"
                    r="14"
                    stroke="currentColor"
                    strokeWidth="2"
                    className="text-destructive/20"
                  />
                  <path
                    d="M16 2 A14 14 0 0 1 30 16"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    className="text-destructive"
                  />
                </svg>
                <svg
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="currentColor"
                  className="relative text-destructive"
                >
                  <rect x="4" y="4" width="16" height="16" rx="3" />
                </svg>
              </button>
            ) : (
              <button
                onClick={handleSend}
                disabled={!hasContent || disabled}
                className="w-8 h-8 flex items-center justify-center rounded-xl bg-primary text-primary-foreground hover:opacity-90 transition-opacity disabled:opacity-30 disabled:cursor-not-allowed"
                title="Send message"
              >
                <svg
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <line x1="22" y1="2" x2="11" y2="13" />
                  <polygon points="22 2 15 22 11 13 2 9 22 2" />
                </svg>
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
