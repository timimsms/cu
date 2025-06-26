package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/auth"
	"github.com/tim/cu/internal/config"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication with ClickUp",
	Long:  `Authenticate cu with ClickUp API using personal tokens or OAuth.`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with ClickUp",
	Long:  `Authenticate with ClickUp using a personal API token or OAuth device flow.`,
	Run: func(cmd *cobra.Command, args []string) {
		token, _ := cmd.Flags().GetString("token")
		workspace, _ := cmd.Flags().GetString("workspace")
		
		authMgr := auth.NewManager()
		
		// If token is provided via flag, use it
		if token != "" {
			authToken := &auth.Token{
				Value:     token,
				Workspace: workspace,
			}
			
			if err := authMgr.SaveToken(workspace, authToken); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save token: %v\n", err)
				os.Exit(1)
			}
			
			fmt.Println("Successfully authenticated!")
			return
		}
		
		// Interactive authentication
		fmt.Println("To authenticate, you'll need a ClickUp personal API token.")
		fmt.Println("You can create one at: https://app.clickup.com/settings/apps")
		fmt.Println()
		
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your ClickUp API token: ")
		tokenInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read token: %v\n", err)
			os.Exit(1)
		}
		
		tokenInput = strings.TrimSpace(tokenInput)
		if tokenInput == "" {
			fmt.Fprintln(os.Stderr, "Token cannot be empty")
			os.Exit(1)
		}
		
		authToken := &auth.Token{
			Value:     tokenInput,
			Workspace: workspace,
		}
		
		if err := authMgr.SaveToken(workspace, authToken); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save token: %v\n", err)
			os.Exit(1)
		}
		
		// Save workspace as default if it's the first one
		if workspace != "" && workspace != auth.DefaultWorkspace {
			config.Set("default_workspace", workspace)
			config.Save()
		}
		
		fmt.Println("\nSuccessfully authenticated!")
		fmt.Println("You can now use cu commands to interact with ClickUp.")
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long:  `Display the current authentication status and user information.`,
	Run: func(cmd *cobra.Command, args []string) {
		authMgr := auth.NewManager()
		workspace := config.GetString("default_workspace")
		if workspace == "" {
			workspace = auth.DefaultWorkspace
		}
		
		token, err := authMgr.GetToken(workspace)
		if err != nil {
			fmt.Println("Not authenticated")
			fmt.Println("\nRun 'cu auth login' to authenticate")
			os.Exit(1)
		}
		
		fmt.Println("Authenticated")
		fmt.Printf("Workspace: %s\n", workspace)
		if token.Email != "" {
			fmt.Printf("Email: %s\n", token.Email)
		}
		fmt.Println("\nToken stored securely in system keychain")
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from ClickUp",
	Long:  `Remove stored authentication credentials.`,
	Run: func(cmd *cobra.Command, args []string) {
		workspace, _ := cmd.Flags().GetString("workspace")
		if workspace == "" {
			workspace = config.GetString("default_workspace")
			if workspace == "" {
				workspace = auth.DefaultWorkspace
			}
		}
		
		authMgr := auth.NewManager()
		if err := authMgr.DeleteToken(workspace); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to logout: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("Successfully logged out from workspace: %s\n", workspace)
	},
}

func init() {
	authLoginCmd.Flags().StringP("token", "t", "", "Personal API token")
	authLoginCmd.Flags().StringP("workspace", "w", "", "Workspace name")
	
	authLogoutCmd.Flags().StringP("workspace", "w", "", "Workspace to logout from")
	
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
}