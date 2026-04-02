package webchatcmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/version"
)

const (
	repoOwner   = "jikime"
	repoName    = "jikime-adk"
	assetPrefix = "jikime-webchat"
)

func newInstallCmd() *cobra.Command {
	var targetVersion string
	var skipBuild bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install or update webchat",
		Long: `Download and install the webchat UI.

Downloads the webchat source from GitHub releases,
then runs pnpm install to set up dependencies.

Examples:
  jikime webchat install              # Install latest version
  jikime webchat install --version 1.7.0
  jikime webchat install --skip-build # Install without building`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Node.js 확인
			if _, err := findNode(); err != nil {
				return err
			}

			// pnpm 확인
			pnpm, err := findPnpm()
			if err != nil {
				return err
			}

			dir := webchatDir()

			// 버전 결정
			ver := targetVersion
			if ver == "" {
				ver = version.String()
			}

			fmt.Printf("  webchat v%s 설치 중...\n", ver)

			// 1. 다운로드
			fmt.Println("  [1/3] 소스 다운로드...")
			if err := downloadAndExtract(ver, dir); err != nil {
				return fmt.Errorf("다운로드 실패: %w", err)
			}

			// 2. pnpm install
			fmt.Println("  [2/3] 의존성 설치 (pnpm install)...")
			install := exec.Command(pnpm, "install", "--frozen-lockfile")
			install.Dir = dir
			install.Stdout = os.Stdout
			install.Stderr = os.Stderr
			if err := install.Run(); err != nil {
				// frozen-lockfile 실패 시 일반 install 재시도
				fmt.Println("  frozen-lockfile 실패, 일반 설치 재시도...")
				retry := exec.Command(pnpm, "install")
				retry.Dir = dir
				retry.Stdout = os.Stdout
				retry.Stderr = os.Stderr
				if err := retry.Run(); err != nil {
					return fmt.Errorf("pnpm install 실패: %w", err)
				}
			}

			// 3. 빌드
			if !skipBuild {
				fmt.Println("  [3/3] 빌드 (pnpm build)...")
				build := exec.Command(pnpm, "build")
				build.Dir = dir
				build.Stdout = os.Stdout
				build.Stderr = os.Stderr
				if err := build.Run(); err != nil {
					return fmt.Errorf("빌드 실패: %w", err)
				}
			} else {
				fmt.Println("  [3/3] 빌드 건너뜀 (--skip-build)")
			}

			fmt.Printf("\n  webchat 설치 완료: %s\n", dir)
			fmt.Println("  실행: jikime webchat start")
			return nil
		},
	}

	cmd.Flags().StringVar(&targetVersion, "version", "", "Install specific version")
	cmd.Flags().BoolVar(&skipBuild, "skip-build", false, "Skip build step after install")

	return cmd
}

// downloadAndExtract downloads the webchat tar.gz from GitHub releases and extracts to destDir.
func downloadAndExtract(ver string, destDir string) error {
	// GitHub release asset URL
	assetName := fmt.Sprintf("%s-v%s.tar.gz", assetPrefix, ver)
	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/%s",
		repoOwner, repoName, ver, assetName)

	// HTTP GET with timeout
	client := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// GitHub token for rate limiting
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if token := os.Getenv("GH_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("다운로드 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 릴리즈가 없으면 로컬 소스 복사 시도
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("릴리즈 v%s에 %s 가 없습니다. GitHub releases를 확인해주세요", ver, assetName)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// 기존 소스 정리 (node_modules, .next 보존)
	if err := cleanSourceFiles(destDir); err != nil {
		return err
	}

	// tar.gz 해제
	return extractTarGz(resp.Body, destDir)
}

// cleanSourceFiles removes source files but preserves node_modules and .next.
func cleanSourceFiles(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	preserve := map[string]bool{
		"node_modules": true,
		".next":        true,
		".env":         true,
		".env.local":   true,
	}

	for _, e := range entries {
		if preserve[e.Name()] {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("%s 삭제 실패: %w", e.Name(), err)
		}
	}

	return nil
}

// extractTarGz extracts a tar.gz stream to destDir.
func extractTarGz(r io.Reader, destDir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip 열기 실패: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar 읽기 실패: %w", err)
		}

		// 경로 보안: 상위 디렉토리 탈출 방지
		clean := filepath.Clean(hdr.Name)
		if strings.HasPrefix(clean, "..") || strings.HasPrefix(clean, "/") {
			continue
		}

		target := filepath.Join(destDir, clean)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			perm := os.FileMode(hdr.Mode) & 0755
			if perm == 0 {
				perm = 0644
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
			if err != nil {
				return err
			}
			// 파일 크기 제한: 100MB
			if _, err := io.Copy(f, io.LimitReader(tr, 100*1024*1024)); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}

// getLatestVersion fetches the latest release version from GitHub API.
func getLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return strings.TrimPrefix(release.TagName, "v"), nil
}
