// Copyright © 2016 Romans Volosatovs <b1101@riseup.net>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	log "github.com/Sirupsen/logrus"
	"github.com/plasma-umass/systemgo/systemctl"
	"github.com/plasma-umass/systemgo/unit"
	"github.com/spf13/cobra"
)

// list-unitsCmd represents the list-units command
var listUnitsCmd = &cobra.Command{
	Use:   "list-units",
	Short: "list units",
	Long:  `list units lists all units known to systemgo`,
	Run: func(cmd *cobra.Command, args []string) {
		var resp systemctl.Response
		if err := client.Call("Server.StatusAll", args, &resp); err != nil {
			log.Error(err)
		}

		if resp.Yield != nil {
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
			fmt.Fprintln(w, "unit\tload\tactive\tsub")
			for name, st := range resp.Yield.(map[string]unit.Status) {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n",
					name, st.Load.Loaded, st.Activation.State, st.Activation.Sub)
			}

			if err := w.Flush(); err != nil {
				log.Error(err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(listUnitsCmd)
}
