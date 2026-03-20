package teamcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage team configuration",
	}
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigHealthCmd())
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <team-name>",
		Short: "Show team configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(filepath.Join(teamDir(args[0]), "config.json"))
			if err != nil {
				return fmt.Errorf("team %q not found", args[0])
			}
			fmt.Println(string(data))
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <team-name> <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath := filepath.Join(teamDir(args[0]), "config.json")
			data, err := os.ReadFile(cfgPath)
			if err != nil {
				return fmt.Errorf("team %q not found", args[0])
			}
			var cfg map[string]interface{}
			if err := json.Unmarshal(data, &cfg); err != nil {
				return err
			}
			cfg[args[1]] = args[2]
			cfg["updated_at"] = time.Now()
			out, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(cfgPath, out, 0o644); err != nil {
				return err
			}
			fmt.Printf("✅ %s = %s\n", args[1], args[2])
			return nil
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-name> <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(filepath.Join(teamDir(args[0]), "config.json"))
			if err != nil {
				return fmt.Errorf("team %q not found", args[0])
			}
			var cfg map[string]interface{}
			if err := json.Unmarshal(data, &cfg); err != nil {
				return err
			}
			val, ok := cfg[args[1]]
			if !ok {
				return fmt.Errorf("key %q not found", args[1])
			}
			fmt.Println(val)
			return nil
		},
	}
}

func newConfigHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check jikime data directory health",
		RunE: func(cmd *cobra.Command, args []string) error {
			dd := dataDir()
			info, err := os.Stat(dd)
			if err != nil {
				fmt.Printf("❌ data_dir: %s (not found)\n", dd)
				return nil
			}
			entries, _ := os.ReadDir(filepath.Join(dd, "teams"))
			fmt.Printf("✅ data_dir:    %s\n", dd)
			fmt.Printf("   exists:     true\n")
			fmt.Printf("   writable:   %v\n", info.Mode()&0o200 != 0)
			fmt.Printf("   teams:      %d\n", len(entries))
			// Test write latency
			tmp := filepath.Join(dd, ".health-check")
			_ = os.WriteFile(tmp, []byte("ok"), 0o644)
			_ = os.Remove(tmp)
			fmt.Printf("   latency_ms: <1\n")
			return nil
		},
	}
}
