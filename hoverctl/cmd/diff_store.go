package cmd

import (
	"fmt"

	"github.com/SpectoLabs/hoverfly/hoverctl/wrapper"
	"github.com/spf13/cobra"
	"github.com/SpectoLabs/hoverfly/core/handlers/v2"
	"bytes"
)

var diffStoreCmd = &cobra.Command{
	Use:   "diff-store",
	Short: "Manage the diffs for Hoverfly",
	Long: `
This allows you to get or clean the differences 
between expected and actual responses stored by 
the Diff mode in Hoverfly. The diffs are represented 
as lists of strings grouped by the same requests.
	`,
}

const errorMsgTemplate = "The \"%s\" parameter is not same - the expected value was [%s], but the actual one [%s]\n"

var getAllDiffStoreCmd = &cobra.Command{
	Use:   "get-all",
	Short: "Gets all diffs stored in Hoverfly",
	Long: `
Returns all differences between expected and actual responses from Hoverfly.
	`,
	Run: func(cmd *cobra.Command, args []string) {

		checkTargetAndExit(target)

		if len(args) == 0 {
			diffs, err := wrapper.GetAllDiffs(*target)
			handleIfError(err)

			output := ""
			for _, diffsWithRequest := range diffs {
				output = fmt.Sprintf("\nFor the request with the simple definition:\n"+
					"\n Method: %s \n Host: %s \n Path: %s \n Query:  %s \n\nhave been recorded %s diff(s):\n",
					diffsWithRequest.Request.Method,
					diffsWithRequest.Request.Host,
					diffsWithRequest.Request.Path,
					diffsWithRequest.Request.Query,
					fmt.Sprint(len(diffsWithRequest.DiffReport)))

				for index, diff := range diffsWithRequest.DiffReport {
					output = output + "\n[" + fmt.Sprint(index+1) + ".]\n" + diffReportMessage(diff) + "\n"
				}
			}

			if output == "" {
				fmt.Println("There are no diffs stored in Hoverfly")
			} else {
				fmt.Println(output)
			}
		}
	},
}

var deleteDiffsCmd = &cobra.Command{
	Use:   "delete-all",
	Short: "Deletes all diffs",
	Long: `
Deletes all differences between expected and actual responses stored in Hoverfly.
	`,
	Run: func(cmd *cobra.Command, args []string) {

		checkTargetAndExit(target)

		err := wrapper.DeleteAllDiffs(*target)
		handleIfError(err)
		fmt.Println("All diffs have been deleted")
	},
}

func diffReportMessage(report v2.DiffReport) string {
	var msg bytes.Buffer
	for index, entry := range report.DiffEntries {
		msg.Write([]byte(fmt.Sprintf("(%d)"+errorMsgTemplate, index+1, entry.Field, entry.Expected, entry.Actual)))
	}
	return msg.String()
}

func init() {
	RootCmd.AddCommand(diffStoreCmd)
	diffStoreCmd.AddCommand(getAllDiffStoreCmd)
	diffStoreCmd.AddCommand(deleteDiffsCmd)
}
