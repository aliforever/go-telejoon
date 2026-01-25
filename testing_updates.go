package telejoon

import (
	"sync"

	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"github.com/aliforever/go-telegram-bot-api/structs"
)

// TestUpdateBuilder helps create test updates for unit testing
type TestUpdateBuilder struct {
	update tgbotapi.Update
}

// NewTestUpdate creates a new test update builder
func NewTestUpdate() *TestUpdateBuilder {
	return &TestUpdateBuilder{
		update: tgbotapi.Update{},
	}
}

// WithMessage adds a message to the update
func (b *TestUpdateBuilder) WithMessage(chatID int64, userID int64, text string) *TestUpdateBuilder {
	b.update.Message = &structs.Message{
		MessageId: 1,
		Text:      text,
		Chat: &structs.Chat{
			Id:   chatID,
			Type: "private",
		},
		From: &structs.User{
			Id: userID,
		},
	}
	return b
}

// WithGroupMessage adds a group message to the update
func (b *TestUpdateBuilder) WithGroupMessage(chatID int64, userID int64, text string) *TestUpdateBuilder {
	b.update.Message = &structs.Message{
		MessageId: 1,
		Text:      text,
		Chat: &structs.Chat{
			Id:   chatID,
			Type: "group",
		},
		From: &structs.User{
			Id: userID,
		},
	}
	return b
}

// WithSuperGroupMessage adds a supergroup message to the update
func (b *TestUpdateBuilder) WithSuperGroupMessage(chatID int64, userID int64, text string) *TestUpdateBuilder {
	b.update.Message = &structs.Message{
		MessageId: 1,
		Text:      text,
		Chat: &structs.Chat{
			Id:   chatID,
			Type: "supergroup",
		},
		From: &structs.User{
			Id: userID,
		},
	}
	return b
}

// WithCallbackQuery adds a callback query to the update
func (b *TestUpdateBuilder) WithCallbackQuery(userID int64, data string) *TestUpdateBuilder {
	b.update.CallbackQuery = &structs.CallbackQuery{
		Id:   "test_callback_id",
		Data: data,
		From: &structs.User{
			Id: userID,
		},
		Message: &structs.Message{
			MessageId: 1,
			Chat: &structs.Chat{
				Id:   userID,
				Type: "private",
			},
		},
	}
	return b
}

// WithUser sets the user details for the update
func (b *TestUpdateBuilder) WithUser(id int64, firstName, username string) *TestUpdateBuilder {
	user := &structs.User{
		Id:        id,
		FirstName: firstName,
		Username:  username,
	}
	if b.update.Message != nil {
		b.update.Message.From = user
	}
	if b.update.CallbackQuery != nil {
		b.update.CallbackQuery.From = user
	}
	return b
}

// WithPhoto adds a photo to the message
func (b *TestUpdateBuilder) WithPhoto(fileID string) *TestUpdateBuilder {
	if b.update.Message == nil {
		b.update.Message = &structs.Message{}
	}
	b.update.Message.Photo = []structs.PhotoSize{
		{FileId: fileID, Width: 100, Height: 100},
	}
	return b
}

// WithDocument adds a document to the message
func (b *TestUpdateBuilder) WithDocument(fileID, fileName string) *TestUpdateBuilder {
	if b.update.Message == nil {
		b.update.Message = &structs.Message{}
	}
	b.update.Message.Document = &structs.Document{
		FileId:   fileID,
		FileName: fileName,
	}
	return b
}

// WithLocation adds a location to the message
func (b *TestUpdateBuilder) WithLocation(latitude, longitude float64) *TestUpdateBuilder {
	if b.update.Message == nil {
		b.update.Message = &structs.Message{}
	}
	b.update.Message.Location = &structs.Location{
		Latitude:  latitude,
		Longitude: longitude,
	}
	return b
}

// WithContact adds a contact to the message
func (b *TestUpdateBuilder) WithContact(phoneNumber, firstName string) *TestUpdateBuilder {
	if b.update.Message == nil {
		b.update.Message = &structs.Message{}
	}
	b.update.Message.Contact = &structs.Contact{
		PhoneNumber: phoneNumber,
		FirstName:   firstName,
	}
	return b
}

// WithMessageID sets a custom message ID
func (b *TestUpdateBuilder) WithMessageID(messageID int64) *TestUpdateBuilder {
	if b.update.Message != nil {
		b.update.Message.MessageId = messageID
	}
	return b
}

// WithReplyTo sets a reply to message
func (b *TestUpdateBuilder) WithReplyTo(messageID int64, text string) *TestUpdateBuilder {
	if b.update.Message != nil {
		b.update.Message.ReplyToMessage = &structs.Message{
			MessageId: messageID,
			Text:      text,
		}
	}
	return b
}

// Build returns the constructed update
func (b *TestUpdateBuilder) Build() tgbotapi.Update {
	return b.update
}

// BuildStateUpdate returns a StateUpdate wrapping the constructed update
func (b *TestUpdateBuilder) BuildStateUpdate(state string) *StateUpdate {
	return &StateUpdate{
		storage: &sync.Map{},
		State:   state,
		Update:  b.update,
	}
}

// BuildStateUpdateWithLanguage returns a StateUpdate with language set
func (b *TestUpdateBuilder) BuildStateUpdateWithLanguage(state string, lang *Language) *StateUpdate {
	return &StateUpdate{
		storage:  &sync.Map{},
		State:    state,
		Update:   b.update,
		language: lang,
	}
}
