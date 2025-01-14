package twitch

import (
	"strings"

	"github.com/nicklaw5/helix/v2"
	"github.com/streampets/backend/models"
)

type TwitchChannel struct {
	Name      string
	UserId    string
	Challenge string
	eventsubs []string
	prefix    string

	JoinChan   chan models.User
	MsgChan    chan models.ChatMessageEvent
	CmdChan    chan models.CommandEvent
	FollowChan chan models.User
	SubChan    chan models.SubEvent
	StreamChan chan models.StreamEvent

	client      *helix.Client
	accessToken string
}

func NewTwitchChannel(channelName, channelUserID, accessToken, challenge, prefix string) (*TwitchChannel, error) {
	ch := TwitchChannel{
		Name:        channelName,
		UserId:      channelUserID,
		Challenge:   challenge,
		prefix:      prefix,
		accessToken: accessToken,
		JoinChan:    make(chan models.User, 10),
		MsgChan:     make(chan models.ChatMessageEvent, 25),
		CmdChan:     make(chan models.CommandEvent, 25),
		FollowChan:  make(chan models.User),
		SubChan:     make(chan models.SubEvent),
	}
	client, err := helix.NewClient(&helix.Options{
		ClientID:        clientId,
		AppAccessToken:  appAccessToken,
		UserAccessToken: ch.accessToken,
	})
	if err != nil {
		return nil, err
	}
	ch.client = client
	for _, v := range EVENT_PATH {
		err := ch.bind(v)
		if err != nil {
			ch.Close()
			return nil, err
		}
	}
	channels[channelUserID] = &ch
	return &ch, nil
}

func (c *TwitchChannel) bind(event string) error {
	req := helix.EventSubSubscription{
		Type:    event,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: c.UserId,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: EVENT_PATH[event],
			Secret:   c.Challenge,
		},
	}
	_, err := c.client.CreateEventSubSubscription(&req)
	if err != nil {
		return err
	}
	c.eventsubs = append(c.eventsubs, req.ID)
	return nil
}

func (c *TwitchChannel) Close() {
	for _, id := range c.eventsubs {
		_, err := c.client.RemoveEventSubSubscription(id)
		if err != nil {
			//TODO I hope Twitch implemented a mechanism to close itself
		}
	}
	clear(c.eventsubs)
	close(c.JoinChan)
	close(c.MsgChan)
	close(c.CmdChan)
	close(c.FollowChan)
	close(c.SubChan)
}

// If the user has never interacted with the stream, sends a JOIN event.
func (c *TwitchChannel) joinOrGet(id models.TwitchID, name string) *models.User {
	var user models.User
	if true { //TODO check from cache or db
		user = models.User{
			UserID:   models.TwitchID(id),
			Username: name,
		}
	} else {
		// get from db
		c.JoinChan <- user
	}
	return &user
}

func (c *TwitchChannel) onMessageReceived(event *helix.EventSubChannelChatMessageEvent) {
	user := *c.joinOrGet(models.TwitchID(event.ChatterUserID), event.ChatterUserName)
	msg, found := strings.CutPrefix(event.Message.Text, c.prefix)
	if !found {
		c.MsgChan <- models.ChatMessageEvent{User: user, Text: msg}
		return
	}
	if args, found := strings.CutPrefix(msg, "JUMP"); found {
		c.CmdChan <- models.CommandEvent{
			User: user, Command: models.CMD_JUMP,
			Args: strings.Split(args, " "),
		}
	} else if args, found := strings.CutPrefix(msg, "COLOR"); found {
		c.CmdChan <- models.CommandEvent{
			User: user, Command: models.CMD_COLOR,
			Args: strings.Split(args, " "),
		}
	} else {
		c.MsgChan <- models.ChatMessageEvent{User: user, Text: msg}
	}
}

func (c *TwitchChannel) onFollow(event *helix.EventSubChannelFollowEvent) {
	c.FollowChan <- *c.joinOrGet(models.TwitchID(event.UserID), event.UserName)
}

func (c *TwitchChannel) onSubscription(event *helix.EventSubChannelSubscribeEvent, isSubbed bool) {
	c.SubChan <- models.SubEvent{
		User:     *c.joinOrGet(models.TwitchID(event.UserID), event.UserName),
		IsSubbed: isSubbed,
	}
}

func (c *TwitchChannel) onStreamStarted(event *helix.EventSubStreamOnlineEvent) {
	c.StreamChan <- models.StreamEvent{
		User:      *c.joinOrGet(models.TwitchID(event.BroadcasterUserID), event.BroadcasterUserName),
		IsOnline:  true,
		Type:      event.Type,
		StartDate: event.StartedAt.Time,
	}
}

func (c *TwitchChannel) onStreamStopped(event *helix.EventSubStreamOfflineEvent) {
	c.StreamChan <- models.StreamEvent{
		User:     *c.joinOrGet(models.TwitchID(event.BroadcasterUserID), event.BroadcasterUserName),
		IsOnline: false,
	}
}
