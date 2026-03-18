"use client"

import * as ResizablePrimitive from "react-resizable-panels"
export { usePanelRef } from "react-resizable-panels"
export type { PanelImperativeHandle } from "react-resizable-panels"

import { cn } from "@/lib/utils"

function ResizablePanelGroup({
  className,
  ...props
}: ResizablePrimitive.GroupProps) {
  return (
    <ResizablePrimitive.Group
      data-slot="resizable-panel-group"
      className={cn(
        "flex h-full w-full aria-[orientation=vertical]:flex-col",
        className
      )}
      {...props}
    />
  )
}

function ResizablePanel({ ...props }: ResizablePrimitive.PanelProps) {
  return <ResizablePrimitive.Panel data-slot="resizable-panel" {...props} />
}

function ResizableHandle({
  withHandle,
  orientation = 'vertical',
  className,
  ...props
}: ResizablePrimitive.SeparatorProps & {
  withHandle?: boolean
  orientation?: 'horizontal' | 'vertical'
}) {
  return (
    <ResizablePrimitive.Separator
      data-slot="resizable-handle"
      className={cn(
        "group relative flex items-center justify-center bg-transparent focus-visible:outline-hidden",
        orientation === 'vertical'
          ? "w-0 after:absolute after:inset-y-0 after:left-1/2 after:w-4 after:-translate-x-1/2"
          : "h-0 w-full after:absolute after:inset-x-0 after:top-1/2 after:h-4 after:-translate-y-1/2",
        className
      )}
      {...props}
    >
      {withHandle && (
        <div className={cn(
          "z-10 shrink-0 rounded-full bg-muted-foreground/20 hover:bg-muted-foreground/40 transition-colors",
          orientation === 'vertical' ? "h-6 w-1" : "h-1 w-6"
        )} />
      )}
    </ResizablePrimitive.Separator>
  )
}

export { ResizableHandle, ResizablePanel, ResizablePanelGroup }
