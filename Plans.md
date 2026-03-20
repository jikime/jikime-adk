# Plans.md

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | ClawTeam 기능 전체를 jikime-adk에 `jikime team` CLI로 통합 — Harness Engineering 병렬화 + 범용 팀 워크플로우 |
| **마일스톤** | Phase 1-3 완료 시 MVP (`jikime team launch` 동작), Phase 4-6 완료 시 Full Release |
| **담당** | 앤써니 오라버니 / Claude |
| **생성일** | 2026-03-18 |

---

## Phase 1: Core Data Layer (internal/team/)

> 팀 시스템의 심장부. 이 Phase 없이는 이후 어떤 것도 동작하지 않습니다.

| Task | 내용 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 1.1  | `internal/team/types.go` — TeamConfig, Task, Message, CostEvent 구조체 정의 | Go 컴파일 성공, 모든 필드 json 태그 포함, go vet 통과 | - | cc:DONE |
| 1.2  | `internal/team/store.go` — Task 원자적 클레임/락/의존성 자동 언블록 구현 | 동시 클레임 테스트 통과, 파일 기반 lock 충돌 없음 | 1.1 | cc:DONE |
| 1.3  | `internal/team/inbox.go` — FIFO 메시지 박스 (fsnotify Watch) 구현 | 메시지 순서 보장 테스트 통과, Watch 이벤트 정상 수신 | 1.1 | cc:DONE |
| 1.4  | `internal/team/registry.go` — Agent 생존 추적 (tmux/PID 체크) 구현 | 죽은 에이전트 30초 이내 감지, 살아있는 에이전트 false positive 없음 | 1.1 | cc:DONE |
| 1.5  | `internal/team/spawner.go` — tmux/subprocess 백엔드 스폰 구현 | tmux 세션 생성/종료 정상 동작, subprocess fallback 동작 | 1.4 | cc:DONE |
| 1.6  | `internal/team/waiter.go` — 팀 완료 대기 폴링 루프 구현 | 모든 태스크 cc:DONE 시 정상 종료, 타임아웃 옵션 동작 | 1.2 | cc:DONE |
| 1.7  | `internal/team/cost.go` — 에이전트별 토큰 비용 추적 구현 | 비용 이벤트 파일 저장 확인, 집계 함수 정확도 검증 | 1.1 | cc:DONE |
| 1.8  | `internal/team/transport.go` — 플러그인 Transport 인터페이스 정의 + 기본 파일 구현 | 인터페이스 컴파일 통과, 파일 Transport로 send/receive 왕복 테스트 통과 | 1.3 | cc:DONE |
| 1.9  | `internal/team/session.go` — SessionStore (세션 저장/복원) 구현 | 세션 직렬화/역직렬화 테스트 통과, `~/.jikime/sessions/` 경로 저장 확인 | 1.1 | cc:DONE |
| 1.10 | `internal/team/plan.go` — PlanManager (플랜 제출/승인/거부) 구현 | 플랜 상태 전이 테스트 통과 (pending→approved/rejected), `~/.jikime/plans/` 저장 확인 | 1.2 | cc:DONE |
| 1.11 | `internal/team/template.go` — TemplateDef 구조체 + Loader 구현 | YAML 템플릿 로드 성공, 필수 필드 유효성 검사 에러 메시지 명확함 | 1.1 | cc:DONE |

---

## Phase 2: CLI 명령 (cmd/teamcmd/)

> `jikime team` 서브커맨드 전체. Phase 1 Core Layer 위에서 동작합니다.

| Task | 내용 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 2.1  | `cmd/teamcmd/` 디렉토리 생성 + `cmd/root.go`에 `team` 서브커맨드 등록 | `jikime team --help` 정상 출력, cobra 명령 구조 등록 확인 | 1.1 | cc:DONE |
| 2.2  | `create`, `spawn`, `status`, `stop` 기본 명령 구현 | 각 명령 `--help` 출력, `create`로 팀 디렉토리 생성 확인 | 2.1 | cc:DONE |
| 2.3  | `config` 명령 (show/set/get/health) 구현 | `jikime team config show` 출력 확인, health 체크 정상 동작 | 2.1 | cc:DONE |
| 2.4  | `identity` 명령 (show/set) 구현 | 에이전트 ID 설정/조회 동작, `~/.jikime/teams/` 경로 반영 확인 | 2.1 | cc:DONE |
| 2.5  | `session` 명령 (save/show/clear) 구현 | 세션 저장/조회/삭제 왕복 테스트 통과 | 1.9, 2.1 | cc:DONE |
| 2.6  | `plan` 명령 (submit/approve/reject) 구현 | 플랜 제출 후 상태 변경 확인, 거부 사유 기록 확인 | 1.10, 2.1 | cc:DONE |
| 2.7  | `lifecycle` 명령 (request-shutdown/approve-shutdown/reject-shutdown/idle/on-exit) 구현 | 셧다운 요청-승인 흐름 동작, idle 상태 전환 확인 | 2.1 | cc:DONE |
| 2.8  | `tasks` 명령 (create/get/update/list/wait) 구현 | 태스크 CRUD 동작, wait 명령으로 완료 대기 확인 | 1.2, 1.6, 2.1 | cc:DONE |
| 2.9  | `inbox` 명령 (send/broadcast/receive/peek/log/watch) 구현 | send→receive 왕복 확인, broadcast 다중 수신 확인, watch 실시간 이벤트 확인 | 1.3, 1.8, 2.1 | cc:DONE |
| 2.10 | `workspace` 명령 (list/checkpoint/merge/cleanup/status) 구현 | git worktree 기반 워크스페이스 생성/목록/병합/정리 동작 확인 | 2.1 | cc:DONE |
| 2.11 | `template` 명령 (list/show) 구현 | YAML 템플릿 목록 출력, 상세 보기 정상 동작 | 1.11, 2.1 | cc:DONE |
| 2.12 | `launch` 명령 — 원클릭 팀 실행 구현 | `jikime team launch --template leader-worker` 로 팀 구성 + 에이전트 스폰 정상 동작 | 1.5, 1.11, 2.2, 2.8 | cc:DONE |
| 2.13 | `board` 명령 (show/live/overview/serve/attach) 구현 | `show`로 Kanban 텍스트 출력, `live`로 실시간 갱신, `serve`로 웹 서버 기동 | 1.2, 1.4, 2.1 | cc:DONE |
| 2.14 | `budget` 명령 (set/show) 구현 | 예산 설정/조회 동작, 초과 시 경고 메시지 출력 확인 | 1.7, 2.1 | cc:DONE |

---

## Phase 3: Hook 통합 (cmd/hookscmd/)

> 기존 hook 시스템에 팀 인식 기능 추가. 에이전트가 팀에 자동으로 합류/이탈합니다.

| Task | 내용 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 3.1  | `cmd/hookscmd/team_agent_start.go` — Join request 전송 + 하트비트 구현 | 에이전트 시작 시 registry에 등록 확인, 하트비트 주기적 갱신 확인 | 1.4, 2.1 | cc:DONE |
| 3.2  | `cmd/hookscmd/team_agent_stop.go` — 태스크 해제 + Cost 기록 구현 | 에이전트 종료 시 클레임 태스크 해제 확인, cost 이벤트 파일 기록 확인 | 1.2, 1.7, 3.1 | cc:DONE |
| 3.3  | `cmd/hookscmd/team_cost_track.go` — PostToolUse 토큰 추적, 예산 초과 시 중단 구현 | 토큰 사용량 누적 추적 확인, 예산 초과 시 에이전트 중단 신호 전송 확인 | 1.7, 3.1 | cc:DONE |
| 3.4  | `cmd/hookscmd/team_plan_gate.go` — Plan → Leader 승인 대기 게이트 구현 | 플랜 제출 후 approved 상태 될 때까지 실행 차단 확인, rejected 시 에러 메시지 출력 | 1.10, 3.1 | cc:DONE |
| 3.5  | `cmd/hookscmd/hooks.go` 수정 — 팀 훅 등록 (JIKIME_TEAM 환경변수 기반 조건부 활성화) | JIKIME_TEAM 설정 시 팀 훅 자동 활성화, 미설정 시 기존 동작 유지 | 3.1, 3.2, 3.3, 3.4 | cc:DONE |

---

## Phase 4: Webchat 백엔드

> 기존 webchat 서버에 팀 관리 API와 SSE 이벤트 스트림을 추가합니다.

| Task | 내용 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 4.1  | `webchat/team.ts` (신규) — TeamFileStore 클래스 구현 | `~/.jikime/teams/` 읽기/쓰기 동작, JSON 파싱 에러 핸들링 확인 | 1.1 설계 기준 | cc:DONE |
| 4.2  | `webchat/team.ts` HTTP 라우트 구현 — 팀 CRUD + 태스크/inbox API | GET/POST/PATCH 엔드포인트 응답 확인, 400/404 에러 응답 적절함 | 4.1 | cc:DONE |
| 4.3  | `webchat/team.ts` SSE 스트림 구현 — 팀 이벤트 실시간 푸시 | SSE 연결 유지 확인, 태스크 상태 변경 이벤트 실시간 수신 확인 | 4.1, 4.2 | cc:DONE |
| 4.4  | `webchat/harness.ts` 수정 — team 모드 분기 추가 | harness.ts 기존 기능 영향 없음, team 모드에서 팀 API 호출로 분기 동작 확인 | 4.2 | cc:DONE |
| 4.5  | `webchat/server.ts` 수정 — `handleTeamRoutes` 등록 | `jikime team board serve` 시 팀 라우트 정상 응답, 기존 라우트 충돌 없음 | 4.3, 4.4 | cc:DONE |

---

## Phase 5: Webchat 프론트엔드

> 팀 대시보드 UI. Kanban 보드와 에이전트 상태를 실시간으로 시각화합니다.

| Task | 내용 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 5.1  | `webchat/src/contexts/TeamContext.tsx` — 팀 상태 컨텍스트 + SSE 구독 구현 | SSE 연결 시 태스크 상태 실시간 갱신 확인, 컨텍스트 Provider 정상 동작 | 4.3 | cc:DONE |
| 5.2  | `webchat/src/components/team/TeamBoard.tsx` — Kanban 보드 컴포넌트 구현 | TODO/WIP/DONE 컬럼 렌더링 확인, 태스크 카드 클릭 시 상세 표시 | 5.1 | cc:DONE |
| 5.3  | `webchat/src/components/team/AgentStatusBar.tsx` — 에이전트 상태 바 구현 | 활성/유휴/오프라인 상태 색상 구분 표시, 하트비트 기준 실시간 갱신 | 5.1 | cc:DONE |
| 5.4  | `webchat/src/components/team/CostMeter.tsx` — 비용 미터 컴포넌트 구현 | 에이전트별 토큰 사용량 막대 표시, 예산 대비 % 표시 | 5.1 | cc:DONE |
| 5.5  | `webchat/src/components/team/TeamCreateModal.tsx` — 팀 생성 모달 구현 | 팀 이름/템플릿 선택 후 생성 API 호출 확인, 생성 성공 시 보드로 이동 | 5.1 | cc:DONE |
| 5.6  | `webchat/src/components/team/TaskAddModal.tsx` — 태스크 추가 모달 구현 | 태스크 제목/DoD/의존성 입력 후 저장 API 호출 확인, 보드에 즉시 반영 | 5.2 | cc:DONE |
| 5.7  | `webchat/src/components/team/MessageInspector.tsx` — 메시지 인스펙터 구현 | inbox 메시지 목록 표시, 메시지 내용 확장 보기 동작 확인 | 5.1 | cc:DONE |
| 5.8  | `webchat/src/components/team/TeamEventLog.tsx` — 팀 이벤트 로그 컴포넌트 구현 | SSE 이벤트 타임라인 형태로 표시, 스크롤 자동 하단 이동 | 5.1 | cc:DONE |
| 5.9  | `webchat/src/components/team/BoardPanel.tsx` — 보드 패널 (전체 팀 대시보드) 구현 | TeamBoard + AgentStatusBar + CostMeter + TeamEventLog 통합 렌더링 확인 | 5.2, 5.3, 5.4, 5.8 | cc:DONE |
| 5.10 | `webchat/src/app/layout` 수정 — board 탭 추가 (기존 탭과 통합) | board 탭 클릭 시 BoardPanel 렌더링 확인, 기존 chat/git 탭 영향 없음 | 5.9 | cc:DONE |

---

## Phase 6: 에이전트 템플릿 + 슬래시 명령

> 실제 에이전트가 역할을 이해하고 행동하도록 안내하는 프롬프트와 사용자 진입점.

| Task | 내용 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 6.1  | `templates/skills/jikime-team-leader.md` — Leader 에이전트 역할 템플릿 작성 | 태스크 분배/승인/모니터링 역할 명확히 기술, `jikime team` 명령 사용법 포함 | 2.12 | cc:DONE |
| 6.2  | `templates/skills/jikime-team-worker.md` — Worker 에이전트 역할 템플릿 작성 | 태스크 클레임/실행/완료 흐름 명확히 기술, `jikime team tasks` 명령 사용법 포함 | 2.8 | cc:DONE |
| 6.3  | `templates/skills/jikime-team-reviewer.md` — Reviewer 에이전트 역할 템플릿 작성 | 결과물 검토/승인/피드백 흐름 명확히 기술, `jikime team plan` 명령 사용법 포함 | 2.6 | cc:DONE |
| 6.4  | `templates/.claude/commands/jikime/team.md` — `/jikime:team` 슬래시 명령 작성 | 슬래시 명령으로 팀 생성~launch까지 원클릭 흐름 안내 동작 확인 | 2.12, 6.1, 6.2 | cc:DONE |
| 6.5  | `templates/.claude/commands/jikime/team-harness.md` — `/jikime:team-harness` 슬래시 명령 작성 | GitHub Issues 기반 병렬 팀 처리 흐름 안내, 기존 harness 명령과 통합 확인 | 6.4 | cc:DONE |
| 6.6  | `templates/.jikime/templates/` — YAML 팀 템플릿 작성 (leader-worker, leader-worker-reviewer, parallel-workers) | YAML 스키마 유효성 검사 통과, `jikime team template list`로 목록 표시 확인 | 1.11, 2.11 | cc:DONE |
| 6.7  | `skills-catalog.yaml` 업데이트 — 신규 팀 스킬 등록 | `jikime skill list`에서 team 관련 스킬 표시 확인 | 6.1, 6.2, 6.3 | cc:DONE |

---

## 데이터 저장 경로 참고

```
~/.jikime/teams/{team-name}/
  team.json          # TeamConfig
  tasks/             # Task 파일들
  inbox/             # 에이전트별 메시지
  events/            # 이벤트 로그
  spawn_registry.json # 살아있는 에이전트 목록
~/.jikime/sessions/{team}/{agent}.json
~/.jikime/costs/{team}/
~/.jikime/plans/
```

---

## 진행 요약

| Phase | 설명 | 태스크 수 | 신규 파일 | 수정 파일 |
|-------|------|-----------|-----------|-----------|
| Phase 1 | Core Data Layer | 11 | 11 | 0 |
| Phase 2 | CLI 명령 | 14 | 14 | 1 |
| Phase 3 | Hook 통합 | 5 | 4 | 1 |
| Phase 4 | Webchat 백엔드 | 5 | 1 | 2 |
| Phase 5 | Webchat 프론트엔드 | 10 | 9 | 1 |
| Phase 6 | 템플릿 + 슬래시 명령 | 7 | 7 | 1 |
| **합계** | | **52** | **~46** | **~6** |
