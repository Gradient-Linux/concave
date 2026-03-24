package cmd

import "github.com/spf13/cobra"

const (
	PresetResearchTeam  = "research-team"
	PresetInferenceNode = "inference-node"
	PresetTrainingNode  = "training-node"
	PresetStudentLab    = "student-lab"
)

var (
	teamName  string
	teamGroup string
	teamUser  string
	teamPreset string
)

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Manage Gradient Linux teams",
}

var teamCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a team",
	RunE:  runScaffoldCommand,
}

var teamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List teams",
	RunE:  runScaffoldCommand,
}

var teamStatusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show team status",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runScaffoldCommand,
}

var teamAddUserCmd = &cobra.Command{
	Use:   "add-user",
	Short: "Add a user to a team",
	RunE:  runScaffoldCommand,
}

var teamRemoveUserCmd = &cobra.Command{
	Use:   "remove-user",
	Short: "Remove a user from a team",
	RunE:  runScaffoldCommand,
}

var teamDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a team",
	RunE:  runScaffoldCommand,
}

func init() {
	teamCreateCmd.Flags().StringVar(&teamName, "name", "", "team name")
	teamCreateCmd.Flags().StringVar(&teamPreset, "preset", "", "team preset")
	teamAddUserCmd.Flags().StringVar(&teamGroup, "group", "", "team group")
	teamAddUserCmd.Flags().StringVar(&teamUser, "user", "", "username")
	teamRemoveUserCmd.Flags().StringVar(&teamGroup, "group", "", "team group")
	teamRemoveUserCmd.Flags().StringVar(&teamUser, "user", "", "username")
	teamDeleteCmd.Flags().StringVar(&teamName, "name", "", "team name")

	teamCmd.AddCommand(teamCreateCmd, teamListCmd, teamStatusCmd, teamAddUserCmd, teamRemoveUserCmd, teamDeleteCmd)
	rootCmd.AddCommand(teamCmd)
}
