// azmeta provides access to the Azure Instance Metadata Service which provides information about running virtual
// machine instances that can be used to manage and configure your virtual machines. This includes information
// such as SKU, network configuration, and upcoming maintenance events. For more information on what type of
// information is available, see metadata APIs.
//
// Azure's Instance Metadata Service is a REST Endpoint accessible to all IaaS VMs created via the Azure Resource
// Manager. The endpoint is available at a well-known non-routable IP address (169.254.169.254) that can be accessed
// only from within the VM.

package azmeta

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest/date"
	"github.com/devigned/tab"
	"github.com/google/uuid"
)

type (
	// Instance is the data structure returned by http://169.254.169.254/metadata/instance
	Instance struct {
		Compute *Compute `json:"compute,omitempty"`
		Network *Network `json:"network,omitempty"`
	}

	// Compute describes the virtual machine details for the instance
	Compute struct {
		AzureEnvironment     string      `json:"azEnvironment,omitempty"`
		CustomData           string      `json:"customData,omitempty"`
		Location             string      `json:"location,omitempty"`
		Name                 string      `json:"name,omitempty"`
		Offer                string      `json:"offer,omitempty"`
		OSType               string      `json:"osType,omitempty"`
		PlacementGroupID     string      `json:"placementGroupId,omitempty"`
		Plan                 *Plan       `json:"plan,omitempty"`
		PlatformFaultDomain  string      `json:"platformFaultDomain,omitempty"`
		PlatformUpdateDomain string      `json:"platformUpdateDomain,omitempty"`
		Provider             string      `json:"provider,omitempty"`
		PublicKeys           []PublicKey `json:"publicKeys,omitempty"`
		Publisher            string      `json:"publisher,omitempty"`
		ResourceGroupName    string      `json:"resourceGroupName,omitempty"`
		ResourceID           string      `json:"resourceId,omitempty"`
		SKU                  string      `json:"sku,omitempty"`
		SubscriptionID       string      `json:"subscriptionId,omitempty"`
		Tags                 string      `json:"tags,omitempty"`
		Version              string      `json:"version,omitempty"`
		VMID                 string      `json:"vmId,omitempty"`
		VMScaleSetName       string      `json:"vmScaleSetName,omitempty"`
		VMSize               string      `json:"vmSize,omitempty"`
		Zone                 string      `json:"zone,omitempty"`
	}

	// PublicKey describes an ssh public key and the path it should be at on the machine
	PublicKey struct {
		KeyData string `json:"keyData,omitempty"`
		Path    string `json:"path,omitempty"`
	}

	// Plan describes the VM Plan
	Plan struct {
		Name      string `json:"name,omitempty"`
		Product   string `json:"product,omitempty"`
		Publisher string `json:"publisher,omitempty"`
	}

	// Network describes the networking details for the instance
	Network struct {
		Interfaces []NetworkInterface `json:"interface,omitempty"`
	}

	// NetworkInterface describes the protocols and addresses for the nic
	NetworkInterface struct {
		IPV4       *Protocol `json:"ipv4,omitempty"`
		IPV6       *Protocol `json:"ipv6,omitempty"`
		MacAddress string    `json:"macAddress,omitemtpy"`
	}

	// Protocol describes the IP Addresses and Subnets
	Protocol struct {
		IPAddresses []Address `json:"ipAddress,omitempty"`
		Subnets     []Subnet  `json:"subnet,omitempty"`
	}

	// Address describes the public and private IP addresses
	Address struct {
		PrivateIPAddress string `json:"privateIpAddress,omitempty"`
		PublicIPAddress  string `json:"publicIpAddress,omitempty"`
	}

	// Subnet describes the subnet for a given protocol
	Subnet struct {
		Address string `json:"address,omitempty"`
		Prefix  string `json:"prefix,omitempty"`
	}

	// Attestation provides a signature and encoding to ensure data is coming from Azure
	Attestation struct {
		Encoding  string `json:"encoding,omitempty"`
		Signature string `json:"signatrue,omitempty"`
	}

	// ScheduledEvents describes a set of events which Azure will execute
	ScheduledEvents struct {
		DocumentIncarnation int              `json:"DocumentIncarnation"`
		Events              []ScheduledEvent `json:"Events"`
	}

	// AckEvents is a set of event ids to be acknowledged so that Azure can complete the event
	AckEvents struct {
		StartRequests []AckEvent `json:"StartRequests"`
	}

	// AckEvent is the event identified for acknowledgement
	AckEvent struct {
		EventID string `json:"EventId"`
	}

	// EventType is the impact this event causes
	//
	// Values:
	//		Freeze: 	The Virtual Machine is scheduled to pause for a few seconds. CPU and network connectivity may be
	//  					suspended, but there is no impact on memory or open files.
	//		Reboot: 	The Virtual Machine is scheduled for reboot (non-persistent memory is lost).
	//		Redeploy: 	The Virtual Machine is scheduled to move to another node (ephemeral disks are lost).
	//		Preempt: 	The Spot Virtual Machine is being deleted (ephemeral disks are lost).
	EventType string

	// EventStatus is the status of the event
	//
	// Values:
	//		Scheduled:	This event is scheduled to start after the time specified in the NotBefore property.
	//		Started:	This event has started.
	//
	// No Completed or similar status is ever provided; the event will no longer be returned when the event is completed.
	EventStatus string

	// ScheduledEvent describes an event which will happen in the future
	ScheduledEvent struct {
		ID           string           `json:"EventID,omitempty"`
		Type         EventType        `json:"EventType,omitempty"`
		ResourceType string           `json:"ResourceType,omitempty"`
		Resources    []string         `json:"Resources,omitempty"`
		Status       EventStatus      `json:"EventStatus,omitempty"`
		NotBefore    date.TimeRFC1123 `json:"NotBefore,omitempty"`
	}

	// IdentityToken is returned by the identity metadata service (basically an AAD JWT)
	//
	// use the access token to auth against Azure services
	IdentityToken struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    string `json:"expires_in"`
		ExpiresOn    string `json:"expires_on"`
		NotBefore    string `json:"not_before"`
		Resource     string `json:"resource"`
		TokenType    string `json:"token_type"`
	}

	// ResourceAndIdentity is the Azure resource ID and the identity to access that resource
	//
	// For more info about Azure resource ids: https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/services-support-managed-identities
	ResourceAndIdentity struct {
		Resource          string     `json:"resource,omitempty"`
		ObjectID          *uuid.UUID `json:"object_id,omitempty"`
		ClientID          *uuid.UUID `json:"client_id,omitempty"`
		ManagedIdentityID *string    `json:"mi_res_id,omitempty"` // Azure resource id
	}

	// Client is the HTTP client for the Cloud Partner Portal
	Client struct {
		HTTPClient                *http.Client
		InstanceAPIVersion        string
		IdentityAPIVersion        string
		ScheduledEventsAPIVersion string
		BaseURI                   string
		mwStack                   []MiddlewareFunc
	}

	// ClientOption is a variadic optional configuration func
	ClientOption func(c *Client) error

	// MiddlewareFunc allows a consumer of the Client to inject handlers within the request / response pipeline
	//
	// The example below adds the atom xml content type to the request, calls the next middleware and returns the
	// result.
	//
	// addAtomXMLContentType MiddlewareFunc = func(next RestHandler) RestHandler {
	//		return func(ctx context.Context, req *http.Request) (res *http.Response, e error) {
	//			if req.Method != http.MethodGet && req.Method != http.MethodHead {
	//				req.Header.Add("content-Type", "application/atom+xml;type=entry;charset=utf-8")
	//			}
	//			return next(ctx, req)
	//		}
	//	}
	MiddlewareFunc func(next RestHandler) RestHandler

	// RestHandler is used to transform a request and response within the http pipeline
	RestHandler func(ctx context.Context, req *http.Request) (*http.Response, error)
)

const (
	// MetadataBaseURI is the local Azure metadata endpoint
	MetadataBaseURI = "http://169.254.169.254/metadata/"

	// InstanceAPIVersion is the highest common API version supported across clouds: https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service#service-availability
	InstanceAPIVersion = "2019-04-30"

	// ScheduledEventsAPIVersion is the highest version of the API at the time
	ScheduledEventsAPIVersion = "2017-11-01"

	// IdentityAPIVersion is the highest version of the Identity API at the time
	IdentityAPIVersion = "2018-02-01"

	// Freeze the Virtual Machine is scheduled to pause for a few seconds. CPU and network connectivity may be
	// suspended, but there is no impact on memory or open files.
	Freeze EventType = "Freeze"
	// Reboot the Virtual Machine is scheduled for reboot (non-persistent memory is lost).
	Reboot EventType = "Reboot"
	// Redeploy the Virtual Machine is scheduled to move to another node (ephemeral disks are lost).
	Redeploy EventType = "Redeploy"
	// Preempt the Spot Virtual Machine is being deleted (ephemeral disks are lost).
	Preempt EventType = "Preempt"

	// Scheduled signifies this event is scheduled to start after the time specified in the NotBefore property.
	Scheduled EventStatus = "Scheduled"
	// Started signifies this event has started.
	Started EventStatus = "Started"
)

var (
	httpLogger MiddlewareFunc = func(next RestHandler) RestHandler {
		return func(ctx context.Context, req *http.Request) (*http.Response, error) {
			printErrLine := func(format string, args ...interface{}) {
				_, _ = fmt.Fprintf(os.Stderr, format, args...)
			}

			requestDump, err := httputil.DumpRequest(req, true)
			if err != nil {
				printErrLine("+%v\n", err)
			}
			printErrLine(string(requestDump))

			res, err := next(ctx, req)
			if err != nil {
				return res, err
			}

			resDump, err := httputil.DumpResponse(res, true)
			if err != nil {
				printErrLine("+%v\n", err)
			}
			printErrLine(string(resDump))

			return res, err
		}
	}

	nonceReg = regexp.MustCompile(`^\d{1,10}$`)
)

// New creates a new Azure Metadata client
func New(opts ...ClientOption) (*Client, error) {
	c := &Client{
		BaseURI:                   MetadataBaseURI,
		InstanceAPIVersion:        InstanceAPIVersion,
		IdentityAPIVersion:        IdentityAPIVersion,
		ScheduledEventsAPIVersion: ScheduledEventsAPIVersion,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// GetAttestation will generate a signed document to verify the data is coming from Azure
//
// nonce is optional and must be digits with a max len of 10; "1234567890"
// https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service#attested-data
func (c *Client) GetAttestation(ctx context.Context, nonce string, middleware ...MiddlewareFunc) (*Attestation, error) {
	path := fmt.Sprintf("attested/document?api-version=%s", c.InstanceAPIVersion)
	if nonce != "" {
		if nonceReg.MatchString(nonce) {
			return nil, fmt.Errorf("nonce must be less than or equal to 10 digits")
		}
		path = fmt.Sprintf("%s&nonce=%s", path, nonce)
	}
	res, err := c.execute(ctx, http.MethodGet, path, nil, middleware...)
	defer closeResponse(ctx, res)

	if err != nil {
		return nil, err
	}

	var attest Attestation
	if err := readAndUnmarshal(res, &attest); err != nil {
		return nil, fmt.Errorf("unable to unmarshal to Attestation: %w", err)
	}

	return &attest, nil
}

// GetScheduledEvents will fetch the scheduled events for the local machine
func (c *Client) GetScheduledEvents(ctx context.Context, middleware ...MiddlewareFunc) (*ScheduledEvents, error) {
	path := fmt.Sprintf("scheduledevents?api-version=%s", c.ScheduledEventsAPIVersion)
	res, err := c.execute(ctx, http.MethodGet, path, nil, middleware...)
	defer closeResponse(ctx, res)

	if err != nil {
		return nil, err
	}

	var se ScheduledEvents
	if err := readAndUnmarshal(res, &se); err != nil {
		return nil, fmt.Errorf("unable to unmarshal to ScheduledEvents: %w", err)
	}

	return &se, nil
}

// AckScheduledEvents will acknowledge a set of scheduled events
func (c *Client) AckScheduledEvents(ctx context.Context, acks AckEvents, middleware ...MiddlewareFunc) error {
	ackJSON, err := json.Marshal(acks)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("scheduledevents?api-version=%s", ScheduledEventsAPIVersion)
	res, err := c.execute(ctx, http.MethodPost, path, bytes.NewReader(ackJSON), middleware...)
	defer closeResponse(ctx, res)
	if err != nil {
		return err
	}

	if res.StatusCode > 299 {
		return fmt.Errorf(fmt.Sprintf("uri: %s, status: %d", res.Request.URL, res.StatusCode))
	}

	return nil
}

// GetInstance will fetch the instance metadata from the local machine
func (c *Client) GetInstance(ctx context.Context, middleware ...MiddlewareFunc) (*Instance, error) {
	path := fmt.Sprintf("instance?api-version=%s", c.InstanceAPIVersion)
	res, err := c.execute(ctx, http.MethodGet, path, nil, middleware...)
	defer closeResponse(ctx, res)

	if err != nil {
		return nil, err
	}

	var instance Instance
	if err := readAndUnmarshal(res, &instance); err != nil {
		return nil, fmt.Errorf("unable to unmarshal to Instance: %w", err)
	}

	return &instance, nil
}

// GetIdentityToken will fetch an authentication token from the instance identity service
func (c *Client) GetIdentityToken(ctx context.Context, tokenReq ResourceAndIdentity, middleware ...MiddlewareFunc) (*IdentityToken, error) {
	if tokenReq.Resource == "" {
		return nil, fmt.Errorf("resource uri must be supplied; see https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/services-support-managed-identities")
	}

	path := fmt.Sprintf("identity/oauth2/token?api-version=%s&resource=%s", c.IdentityAPIVersion, tokenReq.Resource)
	if tokenReq.ManagedIdentityID != nil && (tokenReq.ClientID == nil || tokenReq.ObjectID == nil) {
		return nil, fmt.Errorf("if specifying a managed identity resource id, then client ID and object ID are required")
	}

	if tokenReq.ManagedIdentityID != nil {
		path = fmt.Sprintf("%s&mi_res_id=%s&client_id=%s&object_id=%s", path, *tokenReq.ManagedIdentityID, (*tokenReq.ClientID).String(), (*tokenReq.ObjectID).String())
	}

	res, err := c.execute(ctx, http.MethodGet, path, nil, middleware...)
	defer closeResponse(ctx, res)

	if err != nil {
		return nil, err
	}

	var token IdentityToken
	if err := readAndUnmarshal(res, &token); err != nil {
		return nil, fmt.Errorf("unable to unmarshal to Instance: %w", err)
	}

	return &token, nil
}

func (c *Client) execute(ctx context.Context, method string, entityPath string, body io.Reader, mw ...MiddlewareFunc) (*http.Response, error) {
	req, err := http.NewRequest(method, c.BaseURI+strings.TrimPrefix(entityPath, "/"), body)
	if err != nil {
		tab.For(ctx).Error(err)
		return nil, err
	}

	final := func(_ RestHandler) RestHandler {
		return func(reqCtx context.Context, request *http.Request) (*http.Response, error) {
			client := c.getHTTPClient()
			request = request.WithContext(reqCtx)
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Metadata", "true")
			return client.Do(request)
		}
	}

	mwStack := []MiddlewareFunc{final}
	if os.Getenv("DEBUG") == "true" {
		mwStack = append(mwStack, httpLogger)
	}

	sl := len(c.mwStack) - 1
	for i := sl; i >= 0; i-- {
		mwStack = append(mwStack, c.mwStack[i])
	}

	for i := len(mw) - 1; i >= 0; i-- {
		if mw[i] != nil {
			mwStack = append(mwStack, mw[i])
		}
	}

	var h RestHandler
	for _, mw := range mwStack {
		h = mw(h)
	}

	return h(ctx, req)
}

func (c *Client) getHTTPClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}

	return &http.Client{
		Timeout: 3 * time.Second,
	}
}

func readAndUnmarshal(res *http.Response, v interface{}) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode > 299 {
		return fmt.Errorf(fmt.Sprintf("uri: %s, status: %d, body: %s", res.Request.URL, res.StatusCode, body))
	}

	return json.Unmarshal(body, v)
}

func closeResponse(ctx context.Context, res *http.Response) {
	if res == nil {
		return
	}

	if err := res.Body.Close(); err != nil {
		tab.For(ctx).Error(err)
	}
}
