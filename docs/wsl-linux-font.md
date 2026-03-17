# WSL Rocky Linux 폰트 깨짐 해결 가이드

> 작성일: 2026-03-12
> 환경: Windows WSL + Rocky Linux (RHEL 계열)
> 증상: jikime-adk CLI 실행 시 박스 드로잉 문자, 이모지 깨짐

---

## 원인

jikime-adk는 폰트 파일을 별도로 포함하지 않는다. 대신 Go 소스 코드에 유니코드 특수 문자가 직접 사용된다. 터미널 폰트가 해당 글리프(glyph)를 지원하지 않으면 `?`, `□`, 빈 칸으로 표시된다.

---

## 깨지는 문자 목록

### 박스 드로잉 (배너/UI)

| 위치 | 사용 문자 |
|---|---|
| `cmd/banner/banner.go` | `╔ ╗ ╚ ╝ ║ ═ ╦ ╠ ╣ ╩ ╬ ██ ░▒▓` |
| `cmd/servecmd/serve.go` | `╔ ║ ╚ ═` |
| `cmd/statuscmd/status.go` | `╔ ║ ╚ ═ ─` |
| `cmd/ui/components.go` | `╭ ╮ ╰ ╯ │ ─ ◇` |

### 이모지

| 위치 | 사용 문자 |
|---|---|
| `internal/serve/agent/runner.go` | `🔧 ✓` (채팅 LiveIssueCard에 표시) |
| `cmd/hookscmd/session_start.go` | `🚀` |
| `cmd/hookscmd/user_prompt_submit.go` | `💡` |
| `cmd/statuslinecmd/statusline.go` | `✍️ ⏱ 💬 ● ○ ⟳ ◐ ◑` |
| 템플릿 `.md` 파일 | `✅ ❌ ⚠️ 🔶 🔴` |

---

## 빠른 진단

Rocky Linux에서 아래 명령으로 현재 상태를 확인한다.

```bash
# 로케일 확인
locale | grep LANG

# 박스 드로잉 렌더링 테스트
echo "╔══╗ ║  ║ ╚══╝"

# 이모지 렌더링 테스트
echo "🔧 ✅ ❌ ⚠️"

# 설치된 폰트 목록
fc-list | grep -i "noto\|nerd\|fira"
```

출력이 정상이면 폰트 정상 / 물음표·네모로 나오면 폰트 미설치.

---

## 해결 방법

### 방법 1 — Windows Terminal 폰트 변경 (권장)

WSL을 Windows Terminal에서 사용하는 경우 Windows 쪽 폰트만 변경하면 된다.

**① Nerd Font 다운로드 및 설치 (Windows)**

```
https://www.nerdfonts.com/font-downloads
→ "FiraCode Nerd Font" 또는 "JetBrainsMono Nerd Font" 선택 → 다운로드 → 설치
```

**② Windows Terminal 설정 (`settings.json`)**

```json
{
  "profiles": {
    "defaults": {
      "font": {
        "face": "FiraCode Nerd Font",
        "size": 13
      }
    }
  }
}
```

---

### 방법 2 — Rocky Linux 안에 폰트 설치

SSH, tmux 등 Linux 내부 터미널에서 직접 실행하는 경우.

**① Noto 폰트 (dnf 패키지)**

```bash
sudo dnf install -y google-noto-emoji-fonts google-noto-fonts-common
fc-cache -fv
```

**② Nerd Font 수동 설치**

```bash
mkdir -p ~/.local/share/fonts
cd ~/.local/share/fonts

# FiraCode Nerd Font
curl -LO "https://github.com/ryanoasis/nerd-fonts/releases/latest/download/FiraCode.tar.xz"
tar -xf FiraCode.tar.xz
rm FiraCode.tar.xz
fc-cache -fv

# 설치 확인
fc-list | grep -i fira
```

---

### 방법 3 — 로케일 UTF-8 설정

폰트가 있어도 로케일이 잘못 설정되면 깨질 수 있다.

```bash
# 현재 세션에 즉시 적용
export LANG=en_US.UTF-8
export LC_ALL=en_US.UTF-8

# 영구 적용 (~/.bashrc)
echo 'export LANG=en_US.UTF-8' >> ~/.bashrc
echo 'export LC_ALL=en_US.UTF-8' >> ~/.bashrc
source ~/.bashrc

# 시스템 전체 적용 (root 권한)
sudo localectl set-locale LANG=en_US.UTF-8
```

---

## 우선순위 권장 순서

```
1. 로케일 UTF-8 확인 → 가장 빠른 확인
2. Windows Terminal 폰트 변경 → Windows에서 WSL 사용 시 가장 간단
3. Rocky Linux 내부 폰트 설치 → SSH/headless 환경
```
