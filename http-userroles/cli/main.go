package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cloudmachinery/apps/http-userroles/contracts"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const sockAddr = "/tmp/http-userroles.sock"

func main() {
	client := NewClient(sockAddr)

	root := &cobra.Command{
		Use:          "userroles-cli",
		SilenceUsage: true,
	}

	root.AddCommand(getCommand(client))
	root.AddCommand(createCommand(client))
	root.AddCommand(updateCommand(client))
	root.AddCommand(deleteCommand(client))

	_ = root.Execute()
}

func getCommand(client *Client) *cobra.Command {
	var email, role string

	cmd := &cobra.Command{
		Use: "get",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				users []*contracts.User
				err   error
			)

			switch {
			case email != "" && role != "":
				return errors.New("only one of email or role can be specified")
			case role != "":
				users, err = client.GetUsersByRole(role)
			case email != "":
				user, uerr := client.GetUserByEmail(email)
				if uerr != nil {
					err = uerr
				}
				users = []*contracts.User{user}
			default:
				users, err = client.GetAllUsers()
			}

			if err != nil {
				return err
			}

			tablePrintUsers(users)
			return nil
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address")
	cmd.Flags().StringVarP(&role, "role", "r", "", "Role")

	return cmd
}

func createCommand(client *Client) *cobra.Command {
	var user contracts.User

	cmd := &cobra.Command{
		Use: "create",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.CreateUser(&user); err != nil {
				return err
			}

			switch {
			case user.FullName != "" && len(user.Roles) > 0:
				fmt.Printf("User %q with name %q and roles %v created\n", user.Email, user.FullName, user.Roles)
			case user.FullName != "":
				fmt.Printf("User %q with name %q created\n", user.Email, user.FullName)
			case len(user.Roles) > 0:
				fmt.Printf("User %q with roles %v created\n", user.Email, user.Roles)
			default:
				fmt.Printf("User %q created\n", user.Email)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&user.Email, "email", "e", "", "Email address")
	cmd.Flags().StringVarP(&user.FullName, "name", "n", "", "Full name")
	cmd.Flags().StringSliceVarP(&user.Roles, "roles", "r", []string{}, "Roles")

	return cmd
}

func updateCommand(client *Client) *cobra.Command {
	var user contracts.User

	cmd := &cobra.Command{
		Use: "update",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.UpdateUser(&user); err != nil {
				return err
			}

			fmt.Printf("User %q updated\n", user.Email)
			return nil
		},
	}

	cmd.Flags().StringVarP(&user.Email, "email", "e", "", "Email address")
	cmd.Flags().StringVarP(&user.FullName, "name", "n", "", "Full name")
	cmd.Flags().StringSliceVarP(&user.Roles, "roles", "r", []string{}, "Roles")

	return cmd
}

func deleteCommand(client *Client) *cobra.Command {
	var email string

	cmd := &cobra.Command{
		Use: "delete",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.DeleteUser(email); err != nil {
				return err
			}

			fmt.Printf("User %q deleted\n", email)
			return nil
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address")

	return cmd
}

func tablePrintUsers(users []*contracts.User) {
	if len(users) == 0 {
		fmt.Println("No users found")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Email", "Full Name", "Roles"})
	for _, user := range users {
		table.Append([]string{user.Email, user.FullName, strings.Join(user.Roles, ", ")})
	}
	table.SetFooter([]string{"", "Total", fmt.Sprintf("%d users", len(users))})
	table.Render()
}
