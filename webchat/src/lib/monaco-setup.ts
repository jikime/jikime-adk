import { loader } from '@monaco-editor/react'
import * as monaco from 'monaco-editor'

// Configure @monaco-editor/react to use the locally installed monaco-editor package
// instead of fetching from CDN. This prevents CSP violations and ensures
// the editor works correctly with Next.js Turbopack.
loader.config({ monaco })
