/*
 * Copyright (c) 2018-2019 vChain, Inc. All Rights Reserved.
 * This software is released under GPL3.
 * The full license information can be found under:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 */

package verify

import (
	"fmt"
	"strings"

	"github.com/vchain-us/vcn/pkg/extractor"

	"github.com/vchain-us/vcn/pkg/store"

	"github.com/vchain-us/vcn/pkg/api"
	"github.com/vchain-us/vcn/pkg/meta"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewCmdVerify returns the cobra command for `vcn verify`
func NewCmdVerify() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "verify",
		Example: "  vcn verify /bin/vcn",
		Aliases: []string{"v"},
		Short:   "Verify digital artifact against blockchain",
		Long:    ``,
		RunE:    runVerify,
		Args: func(cmd *cobra.Command, args []string) error {
			if hash, _ := cmd.Flags().GetString("hash"); hash != "" {
				if len(args) > 0 {
					return fmt.Errorf("cannot use arg(s) with --hash")
				}
				return nil
			}
			return cobra.MinimumNArgs(1)(cmd, args)
		},
	}

	cmd.SetUsageTemplate(
		strings.Replace(cmd.UsageTemplate(), "{{.UseLine}}", "{{.UseLine}} ...ARG(s)", 1),
	)

	cmd.Flags().StringSliceP("key", "k", nil, "accept only verification matching the passed key(s)")
	cmd.Flags().String("hash", "", "specify a hash to verify, if set no arg(s) can be used")
	cmd.Flags().StringP("output", "o", "", "output format, one of: --output=json|--output=yaml|--output=''")

	// Bind to VCN_VERIFY_KEYS env var
	viper.BindPFlag("verify_keys", cmd.Flags().Lookup("key"))

	return cmd
}

func runVerify(cmd *cobra.Command, args []string) error {
	hash, err := cmd.Flags().GetString("hash")
	if err != nil {
		return err
	}

	keys := viper.GetStringSlice("verify_keys")

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}
	cmd.SilenceUsage = true

	user := api.NewUser(store.Config().CurrentContext)

	// by hash
	if hash != "" {
		a := &api.Artifact{
			Hash: hash,
		}
		if err := verify(cmd, a, keys, user, output); err != nil {
			return err
		}
		return nil
	}

	// else by args
	for _, arg := range args {
		a, err := extractor.Extract(arg)
		if err != nil {
			return err
		}
		if err := verify(cmd, a, keys, user, output); err != nil {
			return err
		}
	}

	return nil
}

func verify(cmd *cobra.Command, a *api.Artifact, keys []string, user *api.User, output string) (err error) {
	var verification *api.BlockchainVerification
	// if keys have been passed, check for a verification matching them
	if len(keys) > 0 {
		verification, err = api.BlockChainVerifyMatchingPublicKeys(a.Hash, keys)
	} else {
		// if we have an user, check for verification matching user's keys first
		if hasAuth, _ := user.IsAuthenticated(); hasAuth {
			if userKeys := user.Keys(); len(userKeys) > 0 {
				verification, err = api.BlockChainVerifyMatchingPublicKeys(a.Hash, userKeys)
			}
		}
		// if no user nor verification matching the user has found,
		// fallback to the last with highest level avaiable verification
		if verification.Unknown() {
			verification, err = api.BlockChainVerify(a.Hash)
		}
	}

	if err != nil {
		return fmt.Errorf("unable to verify hash: %s", err)
	}

	var artifact *api.ArtifactResponse
	if !verification.Unknown() {
		artifact, _ = api.LoadArtifactForHash(user, a.Hash, verification.MetaHash())
	}

	if err = print(output, a, artifact, verification); err != nil {
		return err
	}

	if output != "" {
		cmd.SilenceErrors = true
	}

	// todo(ameingast): redundant tracking events?
	_ = api.TrackPublisher(user, meta.VcnVerifyEvent)
	_ = api.TrackVerify(user, a.Hash, a.Name)

	if !verification.Trusted() {
		switch true {
		case len(keys) == 1:
			return fmt.Errorf("%s is not verified by %s", a.Hash, keys[0])
		case len(keys) > 1:
			return fmt.Errorf("%s is not verified by any of %s", a.Hash, strings.Join(keys, ", "))
		default:
			return fmt.Errorf("%s is not verified", a.Hash)
		}
	}

	return
}
