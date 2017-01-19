package bluemix

import slsession "github.com/softlayer/softlayer-go/session"

type Config struct {
	Username               string
	SoftlayerApiKey        string
	Password               string
	Region                 string
	Timeout                string
	SoftlayerUsername      string
	SoftlayerEndpointUrl   string
	SoftlayerTimeout       string
	SoftlayerAccountNumber string
}

// ProviderConfig config that contains session
type ProviderConfig interface {
	SoftLayerSession() *slsession.Session
	BluemixSession() *Session
}

type providerConfig struct {
	session *Session
}

// This implements the interface from terraform-provider-softlayer so we can pass in our ProviderConfig
func (config providerConfig) SoftLayerSession() *slsession.Session {
	return config.session.SoftLayerSession
}

// Method to provide the Bluemix Session
func (config providerConfig) BluemixSession() *Session {
	return config.session
}

func (c *Config) Client() (ProviderConfig, error) {

	sess, err := NewSession(c.Username,
		c.Password,
		"",
		c.Region,
		"",
		"",
		c.Timeout,
		c.SoftlayerUsername,
		c.SoftlayerApiKey,
		c.SoftlayerEndpointUrl,
		c.SoftlayerAccountNumber,
		c.SoftlayerTimeout)

	return providerConfig{session: sess}, err
}
