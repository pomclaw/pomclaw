import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeSanitize from 'rehype-sanitize'
import { CodeBlock } from './code-block'

interface MarkdownRendererProps {
  content: string
}

export function MarkdownRenderer({ content }: MarkdownRendererProps) {
  return (
    <div className="text-sm leading-relaxed text-foreground break-words overflow-hidden">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeSanitize]}
        components={{
          code({ className, children, ...props }) {
            const match = /language-(\w+)/.exec(className || '')
            if (!match) {
              return (
                <code
                  className="px-1.5 py-0.5 rounded bg-muted text-accent-foreground font-mono text-[13px]"
                  {...props}
                >
                  {children}
                </code>
              )
            }
            return (
              <CodeBlock
                language={match[1]}
                code={String(children).replace(/\n$/, '')}
              />
            )
          },
          p: ({ children }) => <p className="mb-3 last:mb-0">{children}</p>,
          a: ({ children, href }) => {
            return (
              <a
                href={href}
                className="text-primary hover:underline"
                target="_blank"
                rel="noopener noreferrer"
              >
                {children}
              </a>
            )
          },
          img: ({ src, alt }) => {
            return (
              <img
                src={src}
                alt={alt ?? ''}
                className="max-w-full rounded-lg"
                loading="lazy"
              />
            )
          },
          ul: ({ children }) => (
            <ul className="list-disc ml-5 mb-3 space-y-1">{children}</ul>
          ),
          ol: ({ children }) => (
            <ol className="list-decimal ml-5 mb-3 space-y-1">{children}</ol>
          ),
          li: ({ children }) => <li className="text-sm">{children}</li>,
          blockquote: ({ children }) => (
            <blockquote className="border-l-2 border-primary pl-3 my-3 text-muted-foreground italic">
              {children}
            </blockquote>
          ),
          table: ({ children }) => (
            <div className="overflow-x-auto mb-3">
              <table className="min-w-full text-sm border-collapse">
                {children}
              </table>
            </div>
          ),
          th: ({ children }) => (
            <th className="border border-border px-3 py-1.5 bg-muted text-left font-medium">
              {children}
            </th>
          ),
          td: ({ children }) => (
            <td className="border border-border px-3 py-1.5 text-muted-foreground">
              {children}
            </td>
          ),
          h1: ({ children }) => (
            <h1 className="text-xl font-semibold mb-3 mt-4">{children}</h1>
          ),
          h2: ({ children }) => (
            <h2 className="text-lg font-semibold mb-2 mt-3">{children}</h2>
          ),
          h3: ({ children }) => (
            <h3 className="text-base font-semibold mb-2 mt-3">{children}</h3>
          ),
          hr: () => <hr className="border-border my-4" />,
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}
