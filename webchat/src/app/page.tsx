import { ServerProvider } from '@/contexts/ServerContext'
import { WebSocketProvider } from '@/contexts/WebSocketContext'
import { ProjectProvider } from '@/contexts/ProjectContext'
import AppLayout from '@/components/layout/AppLayout'

export default function Home() {
  return (
    <ServerProvider>
      <WebSocketProvider>
        <ProjectProvider>
          <AppLayout />
        </ProjectProvider>
      </WebSocketProvider>
    </ServerProvider>
  )
}
