package realtime

import "github.com/DannyFestor/go-template-html/internal/domain"

type Authorize func(user domain.User, params map[string]string) bool

type Channels struct {
	rules []channelRule
}

func NewChannels() *Channels {
	return &Channels{rules: nil}
}

func (c *Channels) Public(name string) {
	c.rules = append(c.rules, newChannelRule(name, false, nil))
}

func (c *Channels) Private(pattern string, authorize Authorize) {
	c.rules = append(c.rules, newChannelRule(pattern, true, authorize))
}

func (c *Channels) allowAll(channels []string, user *lazyUser) (bool, error) {
	for _, channel := range channels {
		allowed, err := c.allow(channel, user)
		if err != nil || !allowed {
			return false, err
		}
	}

	return true, nil
}

func (c *Channels) allow(channel string, user *lazyUser) (bool, error) {
	rule, params, found := c.find(channel)
	if !found {
		return false, nil
	}

	if !rule.private {
		return true, nil
	}

	current, signedIn, err := user.get()
	if err != nil {
		return false, err
	}

	return signedIn && current.Verified() && rule.authorize(current, params), nil
}

func (c *Channels) find(channel string) (channelRule, map[string]string, bool) {
	for _, rule := range c.rules {
		params, matches := rule.match(channel)
		if matches {
			return rule, params, true
		}
	}

	var none channelRule

	return none, nil, false
}
