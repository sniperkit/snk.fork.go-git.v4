/*
Sniperkit-Bot
- Date: 2018-08-11 15:40:00.935176804 +0200 CEST m=+0.032827986
- Status: analyzed
*/

package client

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/sniperkit/snk.fork.go-git.v4/plumbing/transport"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type ClientSuite struct{}

var _ = Suite(&ClientSuite{})

func (s *ClientSuite) TestNewClientSSH(c *C) {
	e, err := transport.NewEndpoint("ssh://github.com/sniperkit/snk.fork.go-git.v4")
	c.Assert(err, IsNil)

	output, err := NewClient(e)
	c.Assert(err, IsNil)
	c.Assert(output, NotNil)
}

func (s *ClientSuite) TestNewClientUnknown(c *C) {
	e, err := transport.NewEndpoint("unknown://github.com/sniperkit/snk.fork.go-git.v4")
	c.Assert(err, IsNil)

	_, err = NewClient(e)
	c.Assert(err, NotNil)
}

func (s *ClientSuite) TestNewClientNil(c *C) {
	Protocols["newscheme"] = nil
	e, err := transport.NewEndpoint("newscheme://github.com/sniperkit/snk.fork.go-git.v4")
	c.Assert(err, IsNil)

	_, err = NewClient(e)
	c.Assert(err, NotNil)
}

func (s *ClientSuite) TestInstallProtocol(c *C) {
	InstallProtocol("newscheme", &dummyClient{})
	c.Assert(Protocols["newscheme"], NotNil)
}

func (s *ClientSuite) TestInstallProtocolNilValue(c *C) {
	InstallProtocol("newscheme", &dummyClient{})
	InstallProtocol("newscheme", nil)

	_, ok := Protocols["newscheme"]
	c.Assert(ok, Equals, false)
}

type dummyClient struct {
	*http.Client
}

func (*dummyClient) NewUploadPackSession(*transport.Endpoint, transport.AuthMethod) (
	transport.UploadPackSession, error) {
	return nil, nil
}

func (*dummyClient) NewReceivePackSession(*transport.Endpoint, transport.AuthMethod) (
	transport.ReceivePackSession, error) {
	return nil, nil
}

func typeAsString(v interface{}) string {
	return fmt.Sprintf("%T", v)
}
