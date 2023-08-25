// Copyright 2015 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// Contains the geth command usage template and generator.

package launcher

import (
	"io"
	"sort"

	"github.com/ethereum/go-ethereum/cmd/utils"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/debug"
	"github.com/Fantom-foundation/go-opera/flags"
)

// AppHelpFlagGroups is the application flags, grouped by functionality.
var AppHelpFlagGroups = calcAppHelpFlagGroups()

func calcAppHelpFlagGroups() []flags.FlagGroup {
	overrideFlags()
	overrideParams()

	initFlags()
	return []flags.FlagGroup{
		{
			Name:  "X1",
			Flags: operaFlags,
		},
		{
			Name:  "TRANSACTION POOL",
			Flags: txpoolFlags,
		},
		{
			Name:  "PERFORMANCE TUNING",
			Flags: performanceFlags,
		},
		{
			Name:  "ACCOUNT",
			Flags: accountFlags,
		},
		{
			Name:  "API",
			Flags: rpcFlags,
		},
		{
			Name:  "CONSOLE",
			Flags: consoleFlags,
		},
		{
			Name:  "NETWORKING",
			Flags: networkingFlags,
		},
		{
			Name:  "GAS PRICE ORACLE",
			Flags: gpoFlags,
		},
		{
			Name:  "METRICS AND STATS",
			Flags: metricsFlags,
		},
		{
			Name:  "TESTING",
			Flags: testFlags,
		},
		{
			Name:  "LOGGING AND DEBUGGING",
			Flags: debug.Flags,
		},
		{
			Name:  "ALIASED (deprecated)",
			Flags: legacyRpcFlags,
		},
		{
			Name: "MISC",
			Flags: []cli.Flag{
				cli.HelpFlag,
			},
		},
	}
}

func init() {
	// Override the default app help template
	cli.AppHelpTemplate = flags.AppHelpTemplate

	// Override the default app help printer, but only for the global app help
	originalHelpPrinter := cli.HelpPrinter
	cli.HelpPrinter = func(w io.Writer, tmpl string, data interface{}) {
		if tmpl == flags.AppHelpTemplate {
			// Iterate over all the flags and add any uncategorized ones
			categorized := make(map[string]struct{})
			for _, group := range AppHelpFlagGroups {
				for _, flag := range group.Flags {
					categorized[flag.String()] = struct{}{}
				}
			}
			deprecated := make(map[string]struct{})
			for _, flag := range utils.DeprecatedFlags {
				deprecated[flag.String()] = struct{}{}
			}
			// Only add uncategorized flags if they are not deprecated
			var uncategorized []cli.Flag
			for _, flag := range data.(*cli.App).Flags {
				if _, ok := categorized[flag.String()]; !ok {
					if _, ok := deprecated[flag.String()]; !ok {
						uncategorized = append(uncategorized, flag)
					}
				}
			}
			if len(uncategorized) > 0 {
				// Append all ungategorized options to the misc group
				miscs := len(AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags)
				AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags = append(AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags, uncategorized...)

				// Make sure they are removed afterwards
				defer func() {
					AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags = AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags[:miscs]
				}()
			}
			// Render out custom usage screen
			originalHelpPrinter(w, tmpl, flags.HelpData{App: data, FlagGroups: AppHelpFlagGroups})
		} else if tmpl == flags.CommandHelpTemplate {
			// Iterate over all command specific flags and categorize them
			categorized := make(map[string][]cli.Flag)
			for _, flag := range data.(cli.Command).Flags {
				if _, ok := categorized[flag.String()]; !ok {
					categorized[flags.FlagCategory(flag, AppHelpFlagGroups)] = append(categorized[flags.FlagCategory(flag, AppHelpFlagGroups)], flag)
				}
			}

			// sort to get a stable ordering
			sorted := make([]flags.FlagGroup, 0, len(categorized))
			for cat, flgs := range categorized {
				sorted = append(sorted, flags.FlagGroup{Name: cat, Flags: flgs})
			}
			sort.Sort(flags.ByCategory(sorted))

			// add sorted array to data and render with default printer
			originalHelpPrinter(w, tmpl, map[string]interface{}{
				"cmd":              data,
				"categorizedFlags": sorted,
			})
		} else {
			originalHelpPrinter(w, tmpl, data)
		}
	}
}
