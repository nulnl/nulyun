package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/nulnl/nulyun/settings/webdav"
)

func init() {
	rootCmd.AddCommand(webdavCmd)
}

var webdavCmd = &cobra.Command{
	Use:   "webdav",
	Short: "WebDAV token management command",
	Long:  `Manage WebDAV access tokens, including listing, adding, removing, suspending and activating.`,
	Args:  cobra.NoArgs,
}

func init() {
	webdavCmd.AddCommand(webdavLsCmd)
	webdavCmd.AddCommand(webdavAddCmd)
	webdavCmd.AddCommand(webdavRmCmd)
	webdavCmd.AddCommand(webdavSuspendCmd)
	webdavCmd.AddCommand(webdavActivateCmd)
}

var webdavLsCmd = &cobra.Command{
	Use:   "ls <username>",
	Short: "List all WebDAV tokens for a user",
	Long:  `List all WebDAV access tokens for the specified user.`,
	Args:  cobra.ExactArgs(1),
	RunE: withStore(func(cmd *cobra.Command, args []string, st *store) error {
		username := args[0]

		s, err := st.Settings.Get()
		if err != nil {
			return err
		}

		user, err := st.Users.Get(s.Defaults.Scope, username)
		if err != nil {
			return err
		}

		tokens, err := st.WebDAV.GetByUserID(user.ID)
		if err != nil {
			return err
		}

		if len(tokens) == 0 {
			fmt.Println("This user has no WebDAV tokens")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tName\tToken\tPath\tRead\tWrite\tDelete\tStatus\tCreatedAt")

		for _, token := range tokens {
			displayToken := token.Token
			if len(displayToken) > 16 {
				displayToken = displayToken[:16] + "..."
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%v\t%v\t%v\t%s\t%s\n",
				token.ID,
				token.Name,
				displayToken,
				token.Path,
				token.CanRead,
				token.CanWrite,
				token.CanDelete,
				token.Status,
				token.CreatedAt.Format("2006-01-02 15:04:05"))
		}

		w.Flush()
		return nil
	}, storeOptions{}),
}

var webdavAddCmd = &cobra.Command{
	Use:   "add <username> <name> <path>",
	Short: "Add WebDAV token for a user",
	Long:  `Create a new WebDAV access token for the specified user.`,
	Args:  cobra.ExactArgs(3),
	RunE: withStore(func(cmd *cobra.Command, args []string, st *store) error {
		username := args[0]
		name := args[1]
		path := args[2]

		read, _ := cmd.Flags().GetBool("read")
		write, _ := cmd.Flags().GetBool("write")
		del, _ := cmd.Flags().GetBool("delete")

		s, err := st.Settings.Get()
		if err != nil {
			return err
		}

		user, err := st.Users.Get(s.Defaults.Scope, username)
		if err != nil {
			return err
		}

		token, err := webdav.NewToken(user.ID, name, path, read, write, del)
		if err != nil {
			return err
		}

		err = st.WebDAV.Save(token)
		if err != nil {
			return err
		}

		fmt.Printf("WebDAV token created successfully!\n")
		fmt.Printf("ID: %d\n", token.ID)
		fmt.Printf("Name: %s\n", token.Name)
		fmt.Printf("Token: %s\n", token.Token)
		fmt.Printf("Path: %s\n", token.Path)
		fmt.Printf("Permissions: Read=%v, Write=%v, Delete=%v\n", token.CanRead, token.CanWrite, token.CanDelete)
		fmt.Printf("\nPlease save this token securely; it will only be shown once!\n")
		return nil
	}, storeOptions{}),
}

var webdavRmCmd = &cobra.Command{
	Use:   "rm <username> <id>",
	Short: "Delete WebDAV token",
	Long:  `Delete a WebDAV access token for the specified user.`,
	Args:  cobra.ExactArgs(2),
	RunE: withStore(func(cmd *cobra.Command, args []string, st *store) error {
		username := args[0]

		var id uint
		_, err := fmt.Sscanf(args[1], "%d", &id)
		if err != nil {
			return err
		}

		s, err := st.Settings.Get()
		if err != nil {
			return err
		}

		user, err := st.Users.Get(s.Defaults.Scope, username)
		if err != nil {
			return err
		}

		token, err := st.WebDAV.Get(id)
		if err != nil {
			return err
		}

		if token.UserID != user.ID {
			return fmt.Errorf("token does not belong to user %s", username)
		}

		err = st.WebDAV.Delete(id)
		if err != nil {
			return err
		}

		fmt.Printf("WebDAV token (ID: %d) deleted\n", id)
		return nil
	}, storeOptions{}),
}

var webdavSuspendCmd = &cobra.Command{
	Use:   "suspend <username> <id>",
	Short: "Suspend WebDAV token",
	Long:  `Suspend a WebDAV access token for the specified user.`,
	Args:  cobra.ExactArgs(2),
	RunE: withStore(func(cmd *cobra.Command, args []string, st *store) error {
		username := args[0]

		var id uint
		_, err := fmt.Sscanf(args[1], "%d", &id)
		if err != nil {
			return err
		}

		s, err := st.Settings.Get()
		if err != nil {
			return err
		}

		user, err := st.Users.Get(s.Defaults.Scope, username)
		if err != nil {
			return err
		}

		token, err := st.WebDAV.Get(id)
		if err != nil {
			return err
		}

		if token.UserID != user.ID {
			return fmt.Errorf("token does not belong to user %s", username)
		}

		err = st.WebDAV.Suspend(id)
		if err != nil {
			return err
		}

		fmt.Printf("WebDAV token (ID: %d) suspended\n", id)
		return nil
	}, storeOptions{}),
}

var webdavActivateCmd = &cobra.Command{
	Use:   "activate <username> <id>",
	Short: "Activate WebDAV token",
	Long:  `Activate a WebDAV access token for the specified user.`,
	Args:  cobra.ExactArgs(2),
	RunE: withStore(func(cmd *cobra.Command, args []string, st *store) error {
		username := args[0]

		var id uint
		_, err := fmt.Sscanf(args[1], "%d", &id)
		if err != nil {
			return err
		}

		s, err := st.Settings.Get()
		if err != nil {
			return err
		}

		user, err := st.Users.Get(s.Defaults.Scope, username)
		if err != nil {
			return err
		}

		token, err := st.WebDAV.Get(id)
		if err != nil {
			return err
		}

		if token.UserID != user.ID {
			return fmt.Errorf("token does not belong to user %s", username)
		}

		err = st.WebDAV.Activate(id)
		if err != nil {
			return err
		}

		fmt.Printf("WebDAV token (ID: %d) activated\n", id)
		return nil
	}, storeOptions{}),
}

func init() {
	webdavAddCmd.Flags().Bool("read", true, "read permission")
	webdavAddCmd.Flags().Bool("write", true, "write permission")
	webdavAddCmd.Flags().Bool("delete", false, "delete permission")
}
