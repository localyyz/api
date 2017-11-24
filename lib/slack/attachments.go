package slack

// https://api.slack.com/docs/attachments
// It is possible to create more richly-formatted messages using Attachments.
type AttachmentActionConfirm struct {
	Title       string `json:"title"`
	Text        string `json:"text"`
	OkText      string `json:"ok_text"`
	DismissText string `json:"dismiss_text"`
}

type Option struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}
type SelectedOption Option

type AttachmentAction struct {
	Name    string                   `json:"name"`
	Text    string                   `json:"text"`
	Type    string                   `json:"type"`
	Value   string                   `json:"value,omitempty"`
	Confirm *AttachmentActionConfirm `json:"confirm,omitempty"`
	Style   string                   `json:"style,omitempty"`

	// OptionGroups
	// MinQueryLength

	// valid options:
	// "users"
	// "channels"
	// "conversations"
	DataSource string `json:"data_source"`

	// on callback callback
	Option          []*Option         `json:"options,omitempty"`
	SelectedOptions []*SelectedOption `json:"selected_options,omitempty"`
}

type Attachment struct {
	Color    string `json:"color,omitempty"`
	Fallback string `json:"fallback"`

	AuthorName    string `json:"author_name,omitempty"`
	AuthorSubname string `json:"author_subname,omitempty"`
	AuthorLink    string `json:"author_link,omitempty"`
	AuthorIcon    string `json:"author_icon,omitempty"`

	Title      string `json:"title,omitempty"`
	TitleLink  string `json:"title_link,omitempty"`
	Pretext    string `json:"pretext,omitempty"`
	Text       string `json:"text"`
	CallbackID string `json:"callback_id"`

	ImageURL string `json:"image_url,omitempty"`
	ThumbURL string `json:"thumb_url,omitempty"`

	Footer     string `json:"footer,omitempty"`
	FooterIcon string `json:"footer_icon,omitempty"`
	TimeStamp  int64  `json:"ts,omitempty"`

	Actions    []*AttachmentAction `json:"actions,omitempty"`
	MarkdownIn []string            `json:"mrkdwn_in,omitempty"`
}
