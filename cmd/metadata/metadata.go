package metadata

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/devigned/apmz/pkg/azmeta"
	"github.com/devigned/apmz/pkg/service"
	"github.com/devigned/apmz/pkg/xcobra"
)

type (
	attestArgs struct {
		Nonce string
	}

	tokenArgs struct {
		Resource string
		ObjectID string
		ClientID string
		MIResID  string
	}
)

// NewMetadataCommandGroup will create a new command group for metadata functions
func NewMetadataCommandGroup(sl service.CommandServicer) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:              "metadata",
		Short:            "Azure instance metadata service related commands",
		TraverseChildren: true,
	}

	cmdFuncs := []func(locator service.CommandServicer) (*cobra.Command, error){
		newInstanceCommand,
		newAttestationCommand,
		NewEventsCommandGroup,
		newTokenCommand,
	}

	for _, f := range cmdFuncs {
		cmd, err := f(sl)
		if err != nil {
			return rootCmd, err
		}
		rootCmd.AddCommand(cmd)
	}

	return rootCmd, nil
}

// newInstanceCommand creates a new `apmz metadata instance` command
func newInstanceCommand(sl service.CommandServicer) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "fetches the instance information via the instance metadata service",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			m, err := sl.GetMetadater()
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to build Metadater service: %v", err)
				return err
			}

			instance, err := m.GetInstance(ctx)
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to fetch instance metadata: %v", err)
				return err
			}

			return sl.GetPrinter().Print(instance)
		}),
	}

	return cmd, nil
}

// newAttestationCommand creates a new `apmz metadata attest` command
func newAttestationCommand(sl service.CommandServicer) (*cobra.Command, error) {
	var oArgs attestArgs
	cmd := &cobra.Command{
		Use:   "attest",
		Short: "requests attestation from the metadata instance service",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			m, err := sl.GetMetadater()
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to build Metadater service: %v", err)
				return err
			}

			attest, err := m.GetAttestation(ctx, oArgs.Nonce)
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to fetch attest metadata: %v", err)
				return err
			}

			return sl.GetPrinter().Print(attest)
		}),
	}

	cmd.Flags().StringVarP(&oArgs.Nonce, "nonce", "n", "", "nonce is optional and must be digits with a max len of 10; eg '1234567890'")
	return cmd, nil
}

// newTokenCommand creates a new `apmz metadata token` command
func newTokenCommand(sl service.CommandServicer) (*cobra.Command, error) {
	var oArgs tokenArgs
	cmd := &cobra.Command{
		Use:   "token",
		Short: "requests a token for the System assigned or Managed Identity on the Azure instance",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			if oArgs.MIResID != "" && (oArgs.ClientID == "" || oArgs.ObjectID == "") {
				err := errors.New("when specifying the managed identity resource id, client-id and object-id are required")
				sl.GetPrinter().ErrPrintf("%v", err)
				return err
			}

			m, err := sl.GetMetadater()
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to build Metadater service: %v", err)
				return err
			}

			ri := azmeta.ResourceAndIdentity{
				Resource:          oArgs.Resource,
			}

			if oArgs.MIResID != "" {
				ri.ManagedIdentityID = &oArgs.MIResID
				clientID, err := uuid.Parse(oArgs.ClientID)
				if err != nil {
					sl.GetPrinter().ErrPrintf("client ID could not be parsed into a uuid: %v", err)
					return err
				}
				ri.ClientID = &clientID

				objectID, err := uuid.Parse(oArgs.ObjectID)
				if err != nil {
					sl.GetPrinter().ErrPrintf("object id could not be parsed into a uuid: %v", err)
				}
				ri.ObjectID = &objectID
			}

			token, err := m.GetIdentityToken(ctx, ri)
			if err != nil {
				sl.GetPrinter().ErrPrintf("unable to fetch identity token: %v", err)
				return err
			}

			return sl.GetPrinter().Print(token)
		}),
	}

	cmd.Flags().StringVarP(&oArgs.Resource, "resource", "r", "", "A string indicating the App ID URI of the target resource. It also appears in the aud (audience) claim of the issued token. This example requests a token to access Azure Resource Manager, which has an App ID URI of https://management.azure.com/. For more services and resource IDs see: https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/services-support-managed-identities.")
	cmd.Flags().StringVarP(&oArgs.MIResID, "mi-res", "m", "", "(Optional) Azure resource ID for the managed identity you would like the token for. This is required if the instance has multiple user assigned identities.")
	cmd.Flags().StringVar(&oArgs.ObjectID, "object-id", "", "(Optional) Object ID for the managed identity you would like to use. Required if the VM has multiple user assigned identities.")
	cmd.Flags().StringVar(&oArgs.ClientID, "client-id", "", "(Optional) Client ID for the managed identity you would like to use. Required if the VM has multiple user assigned identities.")
	err := cmd.MarkFlagRequired("resource")
	return cmd, err
}
