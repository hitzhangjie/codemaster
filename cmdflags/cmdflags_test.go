package cmdflags

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func Test_spf13_cmdflags_groupby(t *testing.T) {
	rootCmd := cobra.Command{
		Use:   "root",
		Short: "root cmd description",
		Long:  "root cmd description",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("command args: %v\n", strings.Join(args, " "))
		},
	}

	groups := map[*pflag.FlagSet]string{}
	// group1
	fs1 := pflag.NewFlagSet("features", pflag.ExitOnError)
	fs1.Bool("feature1", true, "enable feature1")
	fs1.Bool("feature2", true, "enable feature2")
	fs1.Bool("feature3", true, "enable feature3")
	rootCmd.Flags().AddFlagSet(fs1)
	groups[fs1] = "features"

	// group2
	fs2 := pflag.NewFlagSet("patches", pflag.ExitOnError)
	fs2.Bool("patch1", true, "apply patch1")
	fs2.Bool("patch2", true, "apply patch2")
	rootCmd.Flags().AddFlagSet(fs2)
	groups[fs2] = "patches"

	// cobra will:
	// - call helpfunc to print help message if defined, otherwise
	// - call usagefunc to print help message if defined, otherwise
	// - call the generated default usagefunc
	//
	// here we use SetUsageFunc is enough.
	rootCmd.SetUsageFunc(func(c *cobra.Command) error {
		for fs, name := range groups {
			usage := fs.FlagUsages()
			idx := strings.IndexFunc(usage, func(r rune) bool {
				return r != ' '
			})
			desc := strings.Repeat(" ", idx) + name + ":"
			help := desc + "\n" + usage
			fmt.Println(help)
		}
		return nil
	})

	// run
	os.Args = []string{"root", "--feature3=false", "--patch2=false", "--help", "helloworld"}

	err := rootCmd.Execute()
	require.Nil(t, err)
}
