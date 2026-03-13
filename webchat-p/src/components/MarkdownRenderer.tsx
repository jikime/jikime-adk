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
      // 기본 텍스트
      'prose-p:leading-relaxed prose-p:my-1.5 prose-p:last:mb-0',
      // 헤딩
      'prose-headings:font-semibold prose-headings:text-zinc-100',
      'prose-h1:text-lg prose-h2:text-base prose-h3:text-sm',
      'prose-headings:mt-4 prose-headings:mb-2 prose-headings:first:mt-0',
      // 인라인 코드
      'prose-code:bg-zinc-700 prose-code:text-emerald-300 prose-code:rounded',
      'prose-code:px-1.5 prose-code:py-0.5 prose-code:text-xs prose-code:font-mono',
      'prose-code:before:content-none prose-code:after:content-none',
      // 코드 블록
      'prose-pre:bg-zinc-900 prose-pre:border prose-pre:border-zinc-700',
      'prose-pre:rounded-lg prose-pre:p-4 prose-pre:my-3',
      'prose-pre:overflow-x-auto',
      // pre 안의 code는 배경 제거
      '[&_pre_code]:bg-transparent [&_pre_code]:text-zinc-200 [&_pre_code]:p-0',
      // 링크
      'prose-a:text-blue-400 prose-a:no-underline hover:prose-a:underline',
      // 리스트
      'prose-ul:my-2 prose-ol:my-2 prose-li:my-0.5',
      'prose-ul:pl-4 prose-ol:pl-4',
      // 인용문
      'prose-blockquote:border-zinc-600 prose-blockquote:text-zinc-400',
      'prose-blockquote:my-2',
      // 구분선
      'prose-hr:border-zinc-700 prose-hr:my-4',
      // 테이블
      'prose-table:text-xs',
      'prose-th:bg-zinc-700 prose-th:text-zinc-200 prose-th:px-3 prose-th:py-1.5',
      'prose-td:px-3 prose-td:py-1.5 prose-td:border-zinc-700',
      'prose-thead:border-zinc-600 prose-tbody:divide-zinc-700',
      // 강조
      'prose-strong:text-zinc-100 prose-em:text-zinc-300',
      className,
    )}>
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        // 코드 블록에 언어 표시
        pre({ children, ...props }) {
          return (
            <pre {...props} className="relative group">
              {children}
            </pre>
          )
        },
        // 코드 블록 내 언어명 표시
        code({ className, children, ...props }) {
          const match = /language-(\w+)/.exec(className || '')
          const isBlock = !!match
          if (isBlock) {
            return (
              <code {...props} className={cn(className, 'block')}>
                {children}
              </code>
            )
          }
          return <code {...props} className={className}>{children}</code>
        },
      }}
    >
      {children}
    </ReactMarkdown>
    </div>
  )
}
