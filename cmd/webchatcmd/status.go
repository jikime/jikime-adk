package webchatcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check webchat installation status",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := webchatDir()

			fmt.Printf("  Location: %s\n", dir)

			// 소스 존재 여부
			pkgPath := filepath.Join(dir, "package.json")
			if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
				fmt.Println("  Status:   Not installed")
				fmt.Println("\n  Run: jikime webchat install")
				return nil
			}

			// 버전 읽기
			data, err := os.ReadFile(pkgPath)
			if err == nil {
				var pkg struct {
					Version string `json:"version"`
				}
				if json.Unmarshal(data, &pkg) == nil && pkg.Version != "" {
					fmt.Printf("  Version:  %s\n", pkg.Version)
				}
			}

			// node_modules 확인
			if isInstalled() {
				fmt.Println("  Deps:     Installed")
			} else {
				fmt.Println("  Deps:     Not installed")
				fmt.Println("\n  Run: jikime webchat install")
				return nil
			}

			// 빌드 확인
			if isBuilt() {
				fmt.Println("  Build:    Ready")
			} else {
				fmt.Println("  Build:    Not built")
				fmt.Println("\n  Run: jikime webchat build")
				return nil
			}

			// Node.js / pnpm 버전
			if node, err := exec.LookPath("node"); err == nil {
				out, _ := exec.Command(node, "--version").Output()
				if len(out) > 0 {
					fmt.Printf("  Node.js:  %s", string(out))
				}
			}
			if pnpm, err := exec.LookPath("pnpm"); err == nil {
				out, _ := exec.Command(pnpm, "--version").Output()
				if len(out) > 0 {
					fmt.Printf("  pnpm:     %s", string(out))
				}
			}

			fmt.Println("\n  Ready! Run: jikime webchat start")
			return nil
		},
	}
}
