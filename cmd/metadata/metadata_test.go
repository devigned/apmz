package metadata

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "github.com/devigned/apmz/internal/test"
	"github.com/devigned/apmz/pkg/azmeta"
)

func TestNewMetadataCommandGroup(t *testing.T) {
	root, err := NewMetadataCommandGroup(nil)
	require.NoError(t, err)

	expected := []string{"instance", "attest", "events", "token"}
	actual := make([]string, len(root.Commands()))
	for i, c := range root.Commands() {
		actual[i] = c.Name()
	}
	assert.ElementsMatch(t, expected, actual)
}

func TestNewAttestationCommand(t *testing.T) {
	cases := []struct {
		name       string
		setup      func(t *testing.T) *mocks.ServiceMock
		assertions func(t *testing.T, cmd *cobra.Command)
	}{
		{
			name: "CommandConstruction",
			setup: func(t *testing.T) *mocks.ServiceMock {
				return nil
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				assert.Equal(t, "attest", cmd.Name())
				if f := cmd.Flags().Lookup("nonce"); assert.NotNil(t, f) {
					assert.Equal(t, "n", f.Shorthand)
				}
			},
		},
		{
			name: "WithoutNonce",
			setup: func(t *testing.T) *mocks.ServiceMock {
				sl := new(mocks.ServiceMock)
				p := new(mocks.PrinterMock)
				m := new(mocks.MetadataMock)
				attest := &azmeta.Attestation{
					Encoding:  "encoding",
					Signature: "sig",
				}
				p.On("Print", attest).Return(nil)
				m.On("GetAttestation", mock.Anything, "", mock.Anything).Return(attest, nil)
				sl.On("GetPrinter").Return(p)
				sl.On("GetMetadater").Return(m, nil)
				return sl
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				assert.NoError(t, cmd.Execute())
			},
		},
		{
			name: "WithNonce",
			setup: func(t *testing.T) *mocks.ServiceMock {
				sl := new(mocks.ServiceMock)
				p := new(mocks.PrinterMock)
				m := new(mocks.MetadataMock)
				attest := &azmeta.Attestation{
					Encoding:  "encoding",
					Signature: "sig",
				}
				p.On("Print", attest).Return(nil)
				m.On("GetAttestation", mock.Anything, "1234567890", mock.Anything).Return(attest, nil)
				sl.On("GetPrinter").Return(p)
				sl.On("GetMetadater").Return(m, nil)
				return sl
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				cmd.SetArgs([]string{"-n", "1234567890"})
				assert.NoError(t, cmd.Execute())
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := c.setup(t)
			cmd, err := NewAttestationCommand(s)
			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			c.assertions(t, cmd)
		})
	}
}

func TestNewTokenCommand(t *testing.T) {
	cases := []struct {
		name       string
		setup      func(t *testing.T) *mocks.ServiceMock
		assertions func(t *testing.T, cmd *cobra.Command)
	}{
		{
			name: "CommandConstruction",
			setup: func(t *testing.T) *mocks.ServiceMock {
				return nil
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				assert.Equal(t, "token", cmd.Name())
				if f := cmd.Flags().Lookup("resource"); assert.NotNil(t, f) {
					assert.Equal(t, "r", f.Shorthand)
				}
				if f := cmd.Flags().Lookup("mi-res"); assert.NotNil(t, f) {
					assert.Equal(t, "m", f.Shorthand)
				}
				assert.NotNil(t, cmd.Flags().Lookup("object-id"))
				assert.NotNil(t, cmd.Flags().Lookup("client-id"))
			},
		},
		{
			name: "WithSystemAssignedIdentity",
			setup: func(t *testing.T) *mocks.ServiceMock {
				sl := new(mocks.ServiceMock)
				p := new(mocks.PrinterMock)
				m := new(mocks.MetadataMock)
				res := azmeta.ResourceAndIdentity{
					Resource: "https://resource.com",
				}
				token := &azmeta.IdentityToken{
					AccessToken:  "foo",
					RefreshToken: "bar",
					NotBefore:    "13434234",
					Resource:     "https://resource.com",
					ExpiresIn:    "675894",
					TokenType:    "jwk",
				}
				p.On("Print", token).Return(nil)
				m.On("GetIdentityToken", mock.Anything, res, mock.Anything).Return(token, nil)
				sl.On("GetPrinter").Return(p)
				sl.On("GetMetadater").Return(m, nil)
				return sl
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				cmd.SetArgs([]string{"-r", "https://resource.com"})
				assert.NoError(t, cmd.Execute())
			},
		},
		{
			name: "WithManagedIdentity",
			setup: func(t *testing.T) *mocks.ServiceMock {
				sl := new(mocks.ServiceMock)
				p := new(mocks.PrinterMock)
				m := new(mocks.MetadataMock)
				objectID := uuid.MustParse("e3dfced6-c558-4602-ac91-2dc4753d18b9")
				clientID := uuid.MustParse("6d16f1dc-06bf-4b26-9915-62ee5680cb76")
				miID := "https://mymanagedidentityid.com/"
				res := azmeta.ResourceAndIdentity{
					Resource:          "https://resource.com",
					ManagedIdentityID: &miID,
					ObjectID:          &objectID,
					ClientID:          &clientID,
				}
				token := &azmeta.IdentityToken{
					AccessToken:  "foo",
					RefreshToken: "bar",
					NotBefore:    "13434234",
					Resource:     "https://resource.com",
					ExpiresIn:    "675894",
					TokenType:    "jwk",
				}
				p.On("Print", token).Return(nil)
				m.On("GetIdentityToken", mock.Anything, res, mock.Anything).Return(token, nil)
				sl.On("GetPrinter").Return(p)
				sl.On("GetMetadater").Return(m, nil)
				return sl
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				cmd.SetArgs([]string{"-r", "https://resource.com", "-m", "https://mymanagedidentityid.com/", "--object-id", "e3dfced6-c558-4602-ac91-2dc4753d18b9", "--client-id", "6d16f1dc-06bf-4b26-9915-62ee5680cb76"})
				assert.NoError(t, cmd.Execute())
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := c.setup(t)
			cmd, err := NewTokenCommand(s)
			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			c.assertions(t, cmd)
		})
	}
}

func TestNewInstanceCommand(t *testing.T) {
	cases := []struct {
		name       string
		setup      func(t *testing.T) *mocks.ServiceMock
		assertions func(t *testing.T, cmd *cobra.Command)
	}{
		{
			name: "CommandConstruction",
			setup: func(t *testing.T) *mocks.ServiceMock {
				return nil
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				assert.Equal(t, "instance", cmd.Name())
			},
		},
		{
			name: "WithSystemAssignedIdentity",
			setup: func(t *testing.T) *mocks.ServiceMock {
				sl := new(mocks.ServiceMock)
				p := new(mocks.PrinterMock)
				m := new(mocks.MetadataMock)

				var instance azmeta.Instance
				js, err := ioutil.ReadFile("./testdata/instance.json")
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(js, &instance))
				p.On("Print", &instance).Return(nil)
				m.On("GetInstance", mock.Anything, mock.Anything).Return(&instance, nil)
				sl.On("GetPrinter").Return(p)
				sl.On("GetMetadater").Return(m, nil)
				return sl
			},
			assertions: func(t *testing.T, cmd *cobra.Command) {
				assert.NoError(t, cmd.Execute())
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := c.setup(t)
			cmd, err := NewInstanceCommand(s)
			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			c.assertions(t, cmd)
		})
	}
}