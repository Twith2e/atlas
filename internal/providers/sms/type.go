package sms

//; // ^ { }  \ [ ~ ] | € ' ”```

type Channel string

const (
	ChannelGeneric  Channel = "generic"  //Used to send promotional messages and messages to phone numbers not on DND (Do Not Disturb).
	ChannelWhatsApp Channel = "whatsapp" //Sends messages via the WhatsApp messaging channel.
	ChannelDND      Channel = "dnd"      //Delivers messages to all phone numbers, regardless of dnd restriction . Ideal for transactional or critical messages.
	ChannelVoice    Channel = "voice"    //	Converts text messages into speech and delivers them as automated voice calls to recipients. Ideal for sending verification codes, alerts, or important notifications through a voice call.
)

type TermiiRequestBody struct {
	APIKey  string  `json:"api_key"`
	To      string  `json:"to"`
	From    string  `json:"from"`
	SMS     string  `json:"sms"`
	Channel Channel `json:"channel"`
	Type    string  `json:"type"`
}
