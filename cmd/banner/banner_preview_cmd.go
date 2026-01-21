package banner

import (
	"github.com/spf13/cobra"
)

// NewBannerPreview creates the banner-preview command
func NewBannerPreview() *cobra.Command {
	return &cobra.Command{
		Use:   "banner-preview",
		Short: "Preview different banner style options for JikiME ADK",
		Long: `Shows 6 different ANSI banner style options:
1. Neon Gradient Style - Cyan to magenta color gradient
2. Isometric 3D Style - 3D perspective effect
3. Glitch/Corrupted Style - Cyberpunk glitch aesthetic
4. Circuit Board Style - Tech/circuit board theme
5. Synthwave Retro Style - 80s sunset gradient
6. Matrix Digital Style - Matrix rain effect`,
		Run: func(cmd *cobra.Command, args []string) {
			PreviewBanners()
		},
	}
}
