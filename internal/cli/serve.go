package cli

import (
	"github.com/Trungsherlock/jobgo/internal/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		mcpMode, _ := cmd.Flags().GetBool("mcp")
		mcpSSE, _ := cmd.Flags().GetBool("mcp-sse")
		port, _ := cmd.Flags().GetInt("port")

		if mcpMode {
			m := server.NewMCP(db)
			return m.ServeStdio()
		}

		if mcpSSE {
			m := server.NewMCP(db)
			return m.ServeSSE(port)
		}

		s := server.New(db)
		return s.ListenAndServe(port)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().Int("port", 8080, "Port to listen on")
	serveCmd.Flags().Bool("mcp", false, "Run as MCP server (stdio)")
	serveCmd.Flags().Bool("mcp-sse", false, "Run as MCP server (HTTP/SSE)")
}
