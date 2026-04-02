package webchatcmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the webchat server",
		Long: `Start the JikiME webchat server.

Requires webchat to be installed first (jikime webchat install).

Examples:
  jikime webchat start
  jikime webchat start --port 3000`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := webchatDir()

			if !isInstalled() {
				return fmt.Errorf("webchat이 설치되지 않았습니다. 먼저 실행: jikime webchat install")
			}

			if !isBuilt() {
				return fmt.Errorf("webchat이 빌드되지 않았습니다. 먼저 실행: jikime webchat build")
			}

			tsx := findTsx(dir)
			if tsx == "" {
				return fmt.Errorf("tsx를 찾을 수 없습니다. webchat을 재설치해주세요: jikime webchat install")
			}

			fmt.Printf("  Starting webchat on port %d...\n", port)
			fmt.Printf("  URL: http://localhost:%d\n\n", port)

			// tsx server.ts 실행
			run := exec.Command(tsx, "server.ts")
			run.Dir = dir
			run.Stdout = os.Stdout
			run.Stderr = os.Stderr
			run.Env = append(os.Environ(),
				"PORT="+strconv.Itoa(port),
				"HOSTNAME=0.0.0.0",
				"NODE_ENV=production",
			)

			// 시그널 전달을 위한 프로세스 그룹
			run.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

			if err := run.Start(); err != nil {
				return fmt.Errorf("webchat 시작 실패: %w", err)
			}

			// Ctrl+C 시그널을 자식 프로세스에 전달
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sig
				if run.Process != nil {
					_ = syscall.Kill(-run.Process.Pid, syscall.SIGTERM)
				}
			}()

			return run.Wait()
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 4000, "Server port")

	return cmd
}
