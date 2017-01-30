/*
* Licensed Materials - Property of IBM
* (C) Copyright IBM Corp. 2017. All Rights Reserved.
* US Government Users Restricted Rights - Use, duplication or
* disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
 */

package bluemix

import slsession "github.com/softlayer/softlayer-go/session"

//Config stores user provider input config
type Config struct {
	Username               string
	SoftlayerAPIKey        string
	Password               string
	Region                 string
	Timeout                string
	SoftlayerUsername      string
	SoftlayerEndpointURL   string
	SoftlayerTimeout       string
	SoftlayerAccountNumber string
}

// ClientSession  contains session Bluemix and Softlayer session
type ClientSession interface {
	SoftLayerSession() *slsession.Session
	BluemixSession() *Session
}

type clientSession struct {
	session *Session
}

// This implements the interface from terraform-provider-softlayer so we can pass in our ProviderConfig
func (sess clientSession) SoftLayerSession() *slsession.Session {
	return sess.session.SoftLayerSession
}

// Method to provide the Bluemix Session
func (sess clientSession) BluemixSession() *Session {
	return sess.session
}

// ClientSession configures and returns a fully initialized ClientSession
func (c *Config) ClientSession() (interface{}, error) {

	sess, err := NewSession(c.Username,
		c.Password,
		"",
		c.Region,
		"",
		"",
		c.Timeout,
		c.SoftlayerUsername,
		c.SoftlayerAPIKey,
		c.SoftlayerEndpointURL,
		c.SoftlayerAccountNumber,
		c.SoftlayerTimeout)

	return clientSession{session: sess}, err
}
