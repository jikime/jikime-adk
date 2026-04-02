package webchatcmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Build webchat for production",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := webchatDir()

			if !isInstalled() {
				return fmt.Errorf("webchat이 설치되지 않았습니다. 먼저 실행: jikime webchat install")
			}

			pnpm, err := findPnpm()
			if err != nil {
				return err
			}

			fmt.Println("  Building webchat...")
			build := exec.Command(pnpm, "build")
			build.Dir = dir
			build.Stdout = os.Stdout
			build.Stderr = os.Stderr
			if err := build.Run(); err != nil {
				return fmt.Errorf("빌드 실패: %w", err)
			}

			fmt.Println("\n  Build complete!")
			fmt.Println("  Run: jikime webchat start")
			return nil
		},
	}
}
