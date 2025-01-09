package twitch

import (
	"errors"
	"github.com/nicklaw5/helix"
)

const (
	BOT_TOKEN = iota
	LRU_LIMIT
	BOT_PREFIX
)

type Channel struct {
	name      string
	challenge string
	userId    string
	eventsubs []string
}

var EVENT_PATH = map[string]string{
	"channel.chat.message":                   "/message",
	helix.EventSubTypeChannelFollow:          "/follow",
	helix.EventSubTypeChannelBan:             "/ban",
	helix.EventSubTypeChannelUnban:           "/ban-end",
	helix.EventSubTypeChannelSubscription:    "/sub",
	helix.EventSubTypeChannelSubscriptionEnd: "/sub-end",
}

var client *helix.Client
var uri string
var channels map[string]*Channel

func Init(URI, clientId, appAccessToken string) {
	var err error = nil
	client, err = helix.NewClient(&helix.Options{
		ClientID:       clientId,
		AppAccessToken: appAccessToken,
	})
	if err != nil {
		panic(err)
	}
	uri = URI + "/wh"
	channels = make(map[string]*Channel)
}

func Close() {
	for _, channel := range channels {
		for _, eventsub := range channel.eventsubs {
			client.RemoveEventSubSubscription(eventsub)
			// TODO Cleanup remaining subscriptions in case of leftovers
		}
	}
}

func (c *Channel) bind(event string) error {
	if _, has := EVENT_PATH[event]; !has {
		return errors.New("No such event")
	}
	_, err := client.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    event,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: c.userId,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: uri + EVENT_PATH[event],
			Secret:   c.challenge,
		},
	})
	if err != nil {
		return err
	}
	//TODO Add subscription-id to subscriptions slice
	return nil
}

func AddChannel(channel, userId, challenge string) error {
	if _, ok := channels[channel]; ok {
		return errors.New("channel already exists")

	}
	channels[channel] = &Channel{
		name:      channel,
		userId:    userId,
		challenge: challenge,
		eventsubs: make([]string, len(EVENT_PATH)),
	}
	for _, v := range EVENT_PATH {
		err := channels[channel].bind(v)
		if err != nil {
			return err
		}
		// TODO Better handling
	}
	return nil
}
