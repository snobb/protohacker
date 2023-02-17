package authority

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"proto/task11/pkg/frame"
)

var authoritiesPerSite = make(map[uint32]*Authority) // site->authority
var mu sync.Mutex

// Authority is a single Authority connection handler
type Authority struct {
	mu        sync.Mutex
	addr      string
	conn      net.Conn
	site      uint32
	policies  map[string]uint32       // species -> policy
	targets   map[string]frame.Target // species -> target
	connected bool
	dialed    bool
	ch        chan *frame.SiteVisit
}

// Handle a single site visit
func HandleSite(ctx context.Context, sv *frame.SiteVisit) error {
	log.Printf("HandleSite: %s", sv)

	auth, err := authority(ctx, sv.Site)
	if err != nil {
		log.Printf("Could not get authority for site: %d", sv.Site)
		frame.WriteError(auth.conn, err)
		return err
	}

	auth.ch <- sv
	return nil
}

// Close the authority connection
func (c *Authority) Close() {
	if c.connected {
		c.connected = false
		c.dialed = false

		if c.ch != nil {
			close(c.ch)
		}

		if c.conn != nil {
			_ = c.conn.Close()
			c.conn = nil
		}

		c.mu.Lock()
		defer mu.Unlock()
		delete(authoritiesPerSite, c.site)
	}
}

func authority(ctx context.Context, site uint32) (*Authority, error) {
	mu.Lock()
	defer mu.Unlock()

	if auth, ok := authoritiesPerSite[site]; ok {
		return auth, nil
	}

	// create new one
	addr := os.Getenv("AUTH_ADDRESS")
	auth := &Authority{
		addr:     addr,
		site:     site,
		policies: make(map[string]uint32),
		targets:  make(map[string]frame.Target),
		ch:       make(chan *frame.SiteVisit),
	}

	if err := auth.dialSite(site); err != nil {
		auth.Close()
		return nil, err
	}

	if err := auth.getPopulations(ctx, site); err != nil {
		auth.Close()
		return nil, err
	}

	go func(auth *Authority) {
		defer auth.Close()

		for sv := range auth.ch {
			if err := auth.handleSiteVisit(ctx, sv); err != nil {
				frame.WriteError(auth.conn, err)
			}
		}
	}(auth)

	authoritiesPerSite[site] = auth

	return auth, nil
}

func (c *Authority) handleSiteVisit(ctx context.Context, sv *frame.SiteVisit) error {
	for name, tgt := range c.targets {
		count, ok := sv.Populations[name]
		if !ok {
			count = 0 // 0 species observed
		}

		// delete policy anyway
		c.mu.Lock()
		if policy, ok := c.policies[name]; ok {
			if err := c.deletePolicy(policy); err != nil {
				log.Printf("handleSiteVisit: deletePolicy(%d): %s", policy, err.Error())
				c.mu.Unlock()
				continue
			}
			delete(c.policies, name)
		}
		c.mu.Unlock()

		var err error
		if count < tgt.Min {
			_, err = c.createPolicy(name, frame.ActionConserve)
		} else if count > tgt.Max {
			_, err = c.createPolicy(name, frame.ActionCull)
		}

		if err != nil {
			frame.WriteError(c.conn, err)
			return err
		}
	}

	return nil
}

func (c *Authority) connect() error {
	if c.connected {
		return nil
	}

	var err error

	log.Print("Connecting to", c.addr)
	c.conn, err = net.Dial("tcp", c.addr)
	if err != nil {
		log.Print("connect: Error client dial:", err.Error())
		return err
	}

	if err := frame.Handshake(c.conn); err != nil {
		log.Print("connect: handshake error", err.Error())
		return err
	}

	c.connected = true

	return nil
}

func (c *Authority) dialSite(site uint32) error {
	if c.dialed && c.site != site {
		return fmt.Errorf("Authority: dialing the second time")
	}

	if err := c.connect(); err != nil {
		return err
	}

	dialAuth := frame.NewDialAuth(site)
	if err := frame.WriteFrame(c.conn, dialAuth); err != nil {
		return err
	}

	c.site = site
	c.dialed = true

	return nil
}

func (c *Authority) getPopulations(ctx context.Context, site uint32) error {
	frm, err := frame.ReadFrame(c.conn)
	if err != nil {
		log.Print("getPopulations: Error recieving frame: ", err.Error())
		return err
	}

	tpops := frame.NewTargetPopulations()
	if err := frm.UnloadInto(tpops); err != nil {
		log.Print("getPopulations: Error parsing frame: ", err.Error())
		return err
	}

	c.targets = tpops.Targets
	return nil
}

func (c *Authority) createPolicy(species string, action frame.Action) (uint32, error) {
	cp := frame.NewCreatePolicy(species, action)
	if err := frame.WriteFrame(c.conn, cp); err != nil {
		log.Printf("createPolicy: Could not send: %v %s: %s",
			species, action, err.Error())
		return 0, err
	}

	frm, err := frame.ReadFrame(c.conn)
	if err != nil {
		log.Print("createPolicy: Error recieving frame: ", err.Error())
		return 0, err
	}

	pr := frame.NewPolicyResult()
	if err := frm.UnloadInto(pr); err != nil {
		log.Print("createPolicy: Error recieving frame: ", err.Error())
		return 0, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.policies[species] = pr.Policy

	return pr.Policy, nil
}

func (c *Authority) deletePolicy(policy uint32) error {
	dp := frame.NewDeletePolicy(policy)
	if err := frame.WriteFrame(c.conn, dp); err != nil {
		log.Print("deletePolicy error:", policy, err.Error())
		return err
	}

	frm, err := frame.ReadFrame(c.conn)
	if err != nil {
		log.Print("deletePolicy: Error recieving frame: ", err.Error())
		return err
	}

	if frm.Kind != frame.KindOK {
		log.Print("deletePolicy: Error recieving frame: ", err.Error())
		return fmt.Errorf("invalid OK response")
	}

	return nil
}
