package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/ignite/cli/v28/ignite/pkg/cliui"
	"github.com/ignite/cli/v28/ignite/pkg/cliui/icons"
	"github.com/ignite/cli/v28/ignite/services/chain"
)

func NewGenerateVuexLegacy() *cobra.Command {
	c := &cobra.Command{
		Use:   "vuex-legacy",
		Short: "TypeScript frontend client and Vuex stores (restored from v0.24.0)",
		Long: `Generate TypeScript frontend client and Vuex stores for your chain.

This command restores the Vuex generation functionality from Ignite CLI v0.24.0.
It generates a TypeScript client along with Vuex store modules for interacting 
with your blockchain from a Vue.js frontend application.

The generated code includes:
- TypeScript client for all modules
- Vuex store modules with actions, mutations, and getters
- Type definitions for all proto messages
- Helper functions for common operations

Example usage:
  ignite g vuex-legacy
  ignite g vuex-legacy --output ./frontend/src/store
`,
		RunE: generateVuexLegacyHandler,
	}

	c.Flags().AddFlagSet(flagSetYes())
	c.Flags().StringP(flagOutput, "o", "", "Vuex store output path")

	return c
}

func generateVuexLegacyHandler(cmd *cobra.Command, _ []string) error {
	session := cliui.New(cliui.StartSpinnerWithText(statusGenerating))
	defer session.End()

	c, err := chain.NewWithHomeFlags(
		cmd,
		chain.WithOutputer(session),
		chain.CollectEvents(session.EventBus()),
		chain.PrintGeneratedPaths(),
	)
	if err != nil {
		return err
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	output, err := cmd.Flags().GetString(flagOutput)
	if err != nil {
		return err
	}

	var opts []chain.GenerateTarget
	if flagGetEnableProtoVendor(cmd) {
		opts = append(opts, chain.GenerateProtoVendor())
	}

	err = c.Generate(cmd.Context(), cacheStorage, chain.GenerateVuexLegacy(output), opts...)
	if err != nil {
		return err
	}

	return session.Println(icons.OK, "Generated TypeScript Client and Vuex stores (legacy)")
}
