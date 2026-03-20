import { ServerProvider } from '@/contexts/ServerContext'
import { WebSocketProvider } from '@/contexts/WebSocketContext'
import { ProjectProvider } from '@/contexts/ProjectContext'
import { TeamProvider } from '@/contexts/TeamContext'
import AppLayout from '@/components/layout/AppLayout'

/**
 * Route Group (app) 레이아웃 — URL에 영향 없이 providers + AppLayout을 공유.
 * / ↔ /project/[...segments] 내비게이션 시 이 레이아웃은 리마운트되지 않으므로
 * Sidebar의 openProjects 등 로컬 상태가 유지된다.
 * children은 의도적으로 렌더링하지 않는다 (AppLayout이 전체 UI를 담당).
 */
export default function AppGroupLayout({ children: _children }: { children: React.ReactNode }) {
  return (
    <ServerProvider>
      <WebSocketProvider>
        <ProjectProvider>
          <TeamProvider>
            <AppLayout />
          </TeamProvider>
        </ProjectProvider>
      </WebSocketProvider>
    </ServerProvider>
  )
}
