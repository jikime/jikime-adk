'use client'

import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { cn } from '@/lib/utils'

interface Props {
  children: string
  className?: string
}

export default function MarkdownRenderer({ children, className }: Props) {
  return (
    <div className={cn(
      'prose prose-invert prose-sm max-w-none break-words',
      'prose-p:leading-relaxed prose-p:my-1.5 prose-p:last:mb-0',
      'prose-headings:font-semibold prose-headings:text-zinc-100',
      'prose-h1:text-lg prose-h2:text-base prose-h3:text-sm',
      'prose-headings:mt-4 prose-headings:mb-2 prose-headings:first:mt-0',
      'prose-code:bg-zinc-700 prose-code:text-emerald-300 prose-code:rounded',
      'prose-code:px-1.5 prose-code:py-0.5 prose-code:text-xs prose-code:font-mono',
      'prose-code:before:content-none prose-code:after:content-none',
      'prose-pre:bg-zinc-900 prose-pre:border prose-pre:border-zinc-700',
      'prose-pre:rounded-lg prose-pre:p-4 prose-pre:my-3 prose-pre:overflow-x-auto',
      '[&_pre_code]:bg-transparent [&_pre_code]:text-zinc-200 [&_pre_code]:p-0',
      'prose-a:text-blue-400 prose-a:no-underline hover:prose-a:underline',
      'prose-ul:my-2 prose-ol:my-2 prose-li:my-0.5 prose-ul:pl-4 prose-ol:pl-4',
      'prose-blockquote:border-zinc-600 prose-blockquote:text-zinc-400 prose-blockquote:my-2',
      'prose-hr:border-zinc-700 prose-hr:my-4',
      'prose-strong:text-zinc-100 prose-em:text-zinc-300',
      className,
    )}>
      <ReactMarkdown remarkPlugins={[remarkGfm]}>
        {children}
      </ReactMarkdown>
    </div>
  )
}
