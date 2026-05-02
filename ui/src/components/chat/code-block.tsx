import { useState } from 'react'
import hljs from 'highlight.js'

interface CodeBlockProps {
  language: string
  code: string
}

export function CodeBlock({ language, code }: CodeBlockProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  // Highlight code using highlight.js
  let highlightedCode: string
  try {
    highlightedCode = hljs.highlight(code, { language }).value
  } catch {
    // Fallback to plain text if language not supported
    highlightedCode = hljs.highlightAuto(code).value
  }

  return (
    <div className="rounded-lg overflow-hidden mb-3 border border-border">
      <div className="flex items-center justify-between px-3 py-1.5 bg-surface-tertiary text-[11px]">
        <span className="text-muted-foreground font-mono">{language}</span>
        <button
          onClick={handleCopy}
          className="text-muted-foreground hover:text-foreground transition-colors"
        >
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>
      <div className="bg-card">
        <pre className="p-3 text-[13px] overflow-x-auto">
          <code
            className={`hljs language-${language}`}
            dangerouslySetInnerHTML={{ __html: highlightedCode }}
          />
        </pre>
      </div>
    </div>
  )
}
