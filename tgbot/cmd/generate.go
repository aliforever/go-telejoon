package cmd

import (
	"fmt"
	"github.com/aliforever/go-telegram-bot-api"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Generator struct {
	BotToken   string
	ModulePath string
	botName    string
}

// NewGenerator creates a new Generator
func NewGenerator(botToken, modulePath string) *Generator {
	return &Generator{
		BotToken:   botToken,
		ModulePath: modulePath,
	}
}

func (g *Generator) Generate() error {
	bot, err := tgbotapi.New(g.BotToken)
	if err != nil {
		return err
	}

	me, err := bot.Send(bot.GetMe())
	if err != nil {
		return err
	}

	g.botName = me.User.Username

	fmt.Printf("Bot Username: %s - ModulePath: %s\n", g.botName, g.ModulePath)

	err = g.createDirectories()
	if err != nil {
		return err
	}

	err = g.createMain()
	if err != nil {
		return err
	}

	err = g.createLangFa()
	if err != nil {
		return err
	}

	err = g.createLangEn()
	if err != nil {
		return err
	}

	err = g.createConfig()
	if err != nil {
		return err
	}

	err = g.createDbRepository()
	if err != nil {
		return err
	}

	err = g.createDbRepositoryUserLanguage()
	if err != nil {
		return err
	}

	err = g.createDbRepositoryUserState()
	if err != nil {
		return err
	}

	err = g.createDbRepositoryUsers()
	if err != nil {
		return err
	}

	err = g.createModelsUser()
	if err != nil {
		return err
	}

	err = g.createModelsUserLanguage()
	if err != nil {
		return err
	}

	err = g.createModelsUserState()
	if err != nil {
		return err
	}

	err = g.createBot()
	if err != nil {
		return err
	}

	err = g.createWelcome()
	if err != nil {
		return err
	}

	err = g.goModInit()
	if err != nil {
		return err
	}

	err = g.goModTidy()
	if err != nil {
		return err
	}

	err = g.goFmt()
	if err != nil {
		return err
	}

	// TODO: need to check if it's installed or not, also need to check if it supports generics
	/*err = g.goSortImports()
	if err != nil {
		return err
	}*/

	return nil
}

func (g *Generator) goModInit() error {
	cmd := exec.Command("go", "mod", "init", g.ModulePath)

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Run()
	if err != nil {
		b, _ := io.ReadAll(errPipe)
		return fmt.Errorf("go mod init failed: %s => %s", err, string(b))
	}

	return nil
}

func (g *Generator) goModTidy() error {
	cmd := exec.Command("go", "mod", "tidy")

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Run()
	if err != nil {
		b, _ := io.ReadAll(errPipe)
		return fmt.Errorf("go mod tidy failed: %s => %s", err, string(b))
	}

	return nil
}

func (g *Generator) goFmt() error {
	cmd := exec.Command("go", "fmt", "./...")

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Run()
	if err != nil {
		b, _ := io.ReadAll(errPipe)
		return fmt.Errorf("go fmt failed: %s => %s", err, string(b))
	}

	return nil
}

func (g *Generator) goSortImports() error {
	cmd := exec.Command("goimports", "-w", ".")

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Run()
	if err != nil {
		b, _ := io.ReadAll(errPipe)
		fmt.Println(string(b))
		return fmt.Errorf("goimports failed: %s => %s", err, string(b))
	}

	return nil
}

func (g *Generator) createDirectories() error {
	err := os.MkdirAll("cmd/bot", os.ModePerm)
	if err != nil {
		return err
	}

	err = os.MkdirAll("lib/bot/config", os.ModePerm)
	if err != nil {
		return err
	}

	err = os.MkdirAll("lib/bot/db", os.ModePerm)
	if err != nil {
		return err
	}

	err = os.MkdirAll("lib/bot/models", os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (g *Generator) createMain() error {
	return os.WriteFile("cmd/bot/main.go", []byte(g.templateMain()), os.ModePerm)
}

func (g *Generator) createLangFa() error {
	return os.WriteFile("locale.fa.toml", []byte(g.templateLangFa()), os.ModePerm)
}

func (g *Generator) createLangEn() error {
	return os.WriteFile("locale.en.toml", []byte(g.templateLangEn()), os.ModePerm)
}

func (g *Generator) createConfig() error {
	return os.WriteFile("lib/bot/config/config.go", []byte(g.templateConfig()), os.ModePerm)
}

func (g *Generator) createDbRepository() error {
	return os.WriteFile("lib/bot/db/repository.go", []byte(g.templateDbRepository()), os.ModePerm)
}

func (g *Generator) createDbRepositoryUserLanguage() error {
	return os.WriteFile(
		"lib/bot/db/userlanguage.go",
		[]byte(g.templateDbRepositoryUserLanguage()),
		os.ModePerm,
	)
}

func (g *Generator) createDbRepositoryUserState() error {
	return os.WriteFile(
		"lib/bot/db/userstate.go",
		[]byte(g.templateDbRepositoryUserState()),
		os.ModePerm,
	)
}

func (g *Generator) createDbRepositoryUsers() error {
	return os.WriteFile("lib/bot/db/users.go", []byte(g.templateDbRepositoryUsers()), os.ModePerm)
}

func (g *Generator) createModelsUser() error {
	return os.WriteFile("lib/bot/models/user.go", []byte(g.templateModelsUser()), os.ModePerm)
}

func (g *Generator) createModelsUserLanguage() error {
	return os.WriteFile("lib/bot/models/userlanguage.go", []byte(g.templateModelsUserLanguage()), os.ModePerm)
}

func (g *Generator) createModelsUserState() error {
	return os.WriteFile("lib/bot/models/userstate.go", []byte(g.templateModelsUserState()), os.ModePerm)
}

func (g *Generator) createBot() error {
	return os.WriteFile("lib/bot/bot.go", []byte(g.templateBot()), os.ModePerm)
}

func (g *Generator) createWelcome() error {
	return os.WriteFile("lib/bot/welcome.go", []byte(g.templateWelcome()), os.ModePerm)
}

// templateMain is the template for cmd/bot/main.go file
func (g *Generator) templateMain() string {
	tpl := `package main

import (
	"context"
	"github.com/aliforever/go-telegram-bot-api"
	"github.com/aliforever/go-telejoon"
	"{{MODULE_PATH}}/lib/bot"
	"{{MODULE_PATH}}/lib/bot/config"
	"{{MODULE_PATH}}/lib/bot/db"
	"github.com/caarlos0/env/v8"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/text/language"
	"time"
)

func main() {
	var cfg config.Config

	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.Level(cfg.LogLevel))

	logger.Infof("Running bot with config: %+v", cfg)

	botAPI, err := tgbotapi.New(cfg.BotToken)
	if err != nil {
		logger.Panic(err)
	}

	if cfg.LogGroupID != 0 {
		logger.AddHook(botAPI.LogrusPeriodicLogger(cfg.LogGroupID, time.Minute*1, "bot"))
	}

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Mongo.Uri))
	if err != nil {
		logger.Panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		logger.Panic(err)
	}

	repo := db.NewRepository(mongoClient.Database(cfg.Mongo.Name), redisClient)

	languages, err := telejoon.NewLanguageBuilder(language.English).
		RegisterTomlFormat([]string{
			"locale.en.toml",
			"locale.fa.toml",
		}).Build()
	if err != nil {
		logger.Panic(err)
	}

	go bot.NewBot(botAPI, repo, languages, cfg.LogGroupID, logger).Start()

	logger.Panic(botAPI.GetUpdates().LongPoll())
}`

	return g.replaceModulePath(tpl)
}

// templateLangFa is the template for cmd/bot/langs/locale.fa.toml file
func (g *Generator) templateLangFa() string {
	tpl := `[General]
BackButton = "➡️ بازگشت"

[ChooseLanguage]
Button = "🇮🇷 فارسی"
Text = "🇮🇷 زبان خود را انتخاب کنید"

[Welcome]
Main = "خوش آمدید!"
ChangeLanguageButton = "🇮🇷 تغییر زبان"`

	return tpl
}

// templateLangEn is the template for cmd/bot/langs/locale.en.toml file
func (g *Generator) templateLangEn() string {
	tpl := `[General]
BackButton = "⬅️ Back"

[ChooseLanguage]
Button = "🇺🇸 English"
Text = "🇺🇸 Choose your language"

[Welcome]
Main = "Welcome"
ChangeLanguageButton = "🇺🇸 Switch Language"`

	return tpl
}

// templateConfig is the template for lib/bot/config/config.go file
func (g *Generator) templateConfig() string {
	tpl := `package config

type Config struct {
	BotToken   string ____env:"BOT_TOKEN" envDefault:"{{BOT_TOKEN}}"____
	LogGroupID int64  ____env:"LOG_GROUP_ID"____
	LogLevel   int    ____env:"LOG_LEVEL" envDefault:"6"____
	Mongo      Mongo  ____envPrefix:"MONGO_"____
	Redis      Redis  ____envPrefix:"REDIS_"____
}

type Mongo struct {
	Uri  string ____env:"URI" envDefault:"mongodb://localhost:27017"____
	Name string ____env:"NAME" envDefault:"{{BOT_NAME}}"____
}

type Redis struct {
	Address  string ____env:"ADDRESS" envDefault:"localhost:6379"____
	Password string ____env:"PASSWORD"____
	DB       int    ____env:"DB" envDefault:"0"____
}
`

	tpl = strings.ReplaceAll(tpl, "____", "`")
	tpl = strings.ReplaceAll(tpl, "{{BOT_TOKEN}}", g.BotToken)

	return g.replaceBotName(tpl)
}

// templateDbRepository is the template for lib/bot/db/repository.go file
func (g *Generator) templateDbRepository() string {
	tpl := `package db

import (
	"github.com/aliforever/go-telegram-bot-api/structs"
	"{{MODULE_PATH}}/lib/bot/models"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	Users        Users
	UserState    UserState
	UserLanguage UserLanguage
}

func NewRepository(db *mongo.Database, rc *redis.Client) *Repository {
	return &Repository{
		Users:        NewMongoUsers(db),
		UserState:    NewMongoUserState(db),
		UserLanguage: NewMongoUserLanguage(db),
	}
}

func (r *Repository) UpsertUser(user *structs.User) error {
	var (
		lastName *string
		username *string
	)

	if user.LastName != "" {
		lastName = &user.LastName
	}

	if user.Username != "" {
		username = &user.Username
	}

	return r.Users.Upsert(&models.User{
		ID:        user.Id,
		Firstname: user.FirstName,
		Lastname:  lastName,
		Username:  username,
	})
}

func (r *Repository) SetUserState(id int64, state string) error {
	return r.UserState.Upsert(&models.UserState{
		ID:    id,
		State: state,
	})
}

func (r *Repository) GetUserState(id int64) (string, error) {
	us, err := r.UserState.FindByID(id)
	if err != nil {
		return "", err
	}

	return us.State, nil
}

func (r *Repository) SetUserLanguage(userID int64, languageTag string) error {
	_ = r.UserLanguage.Upsert(&models.UserLanguage{
		ID:  userID,
		Tag: languageTag,
	})

	return nil
}

func (r *Repository) GetUserLanguage(userID int64) (string, error) {
	ul, err := r.UserLanguage.FindByID(userID)
	if err != nil {
		return "", err
	}

	return ul.Tag, nil
}`

	return g.replaceModulePath(tpl)
}

// templateDbRepositoryUserLanguage is the template for lib/bot/db/repository/userlanguage.go file
func (g *Generator) templateDbRepositoryUserLanguage() string {
	tpl := `package db

import (
	"github.com/aliforever/go-mongolio"
	"{{MODULE_PATH}}/lib/bot/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserLanguage interface {
	// FindByID returns the language of the user
	FindByID(id int64) (*models.UserLanguage, error)
	Upsert(user *models.UserLanguage) error
}

type mongoUserLanguage struct {
	model *mongolio.C[models.UserLanguage]
}

func NewMongoUserLanguage(db *mongo.Database) UserLanguage {
	return &mongoUserLanguage{model: mongolio.Collection[models.UserLanguage](db, "user_language")}
}

func (m mongoUserLanguage) FindByID(id int64) (*models.UserLanguage, error) {
	return m.model.FindByID(id)
}

func (m mongoUserLanguage) Insert(user *models.UserLanguage) error {
	_, err := m.model.Insert(user)

	return err
}

func (m mongoUserLanguage) Upsert(user *models.UserLanguage) error {
	_, err := m.model.UpdateByID(user.ID, bson.M{
		"tag": user.Tag,
	}, options.Update().SetUpsert(true))

	return err
}`

	return g.replaceModulePath(tpl)
}

// templateDbRepositoryUserState is the template for lib/bot/db/repository/userstate.go file
func (g *Generator) templateDbRepositoryUserState() string {
	tpl := `package db

import (
	"github.com/aliforever/go-mongolio"
	"{{MODULE_PATH}}/lib/bot/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserState interface {
	// FindByID returns the state of the user
	FindByID(id int64) (*models.UserState, error)
	Upsert(user *models.UserState) error
}

type mongoUserState struct {
	model *mongolio.C[models.UserState]
}

func NewMongoUserState(db *mongo.Database) UserState {
	return &mongoUserState{model: mongolio.Collection[models.UserState](db, "user_state")}
}

func (m mongoUserState) FindByID(id int64) (*models.UserState, error) {
	return m.model.FindByID(id)
}

func (m mongoUserState) Insert(user *models.UserState) error {
	_, err := m.model.Insert(user)

	return err
}

func (m mongoUserState) Upsert(user *models.UserState) error {
	_, err := m.model.UpdateByID(user.ID, bson.M{
		"state": user.State,
	}, options.Update().SetUpsert(true))

	return err
}`

	return g.replaceModulePath(tpl)
}

// templateDbRepositoryUsers is the template for lib/bot/db/repository/users.go file
func (g *Generator) templateDbRepositoryUsers() string {
	tpl := `package db

import (
	"github.com/aliforever/go-mongolio"
	"{{MODULE_PATH}}/lib/bot/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
)

type Users interface {
	FindById(id int64) (*models.User, error)
	Upsert(user *models.User) error
}

type mongoUsers struct {
	model *mongolio.C[models.User]

	locker sync.Map
}

func NewMongoUsers(db *mongo.Database) Users {
	return &mongoUsers{model: mongolio.Collection[models.User](db, "users")}
}

func (m *mongoUsers) FindById(id int64) (*models.User, error) {
	result, _ := m.locker.LoadOrStore(id, &sync.Mutex{})
	result.(*sync.Mutex).Lock()
	defer result.(*sync.Mutex).Unlock()

	return m.model.FindByID(id)
}

func (m *mongoUsers) Upsert(user *models.User) error {
	_, err := m.model.UpdateByID(user.ID, bson.M{
		"firstname": user.Firstname,
		"lastname":  user.Lastname,
		"username":  user.Username,
	}, options.Update().SetUpsert(true))

	return err
}`

	return g.replaceModulePath(tpl)
}

// templateModelsUser is the template for lib/bot/models/user.go file
func (g *Generator) templateModelsUser() string {
	tpl := `package models

type User struct {
	ID        int64   ___bson:"_id"___
	Firstname string  ___bson:"firstname"___
	Lastname  *string ___bson:"lastname"___
	Username  *string ___bson:"username"___
}`

	tpl = strings.ReplaceAll(tpl, "___", "`")

	return g.replaceModulePath(tpl)
}

// templateModelsUserLanguage is the template for lib/bot/models/userlanguage.go file
func (g *Generator) templateModelsUserLanguage() string {
	tpl := `package models

type UserLanguage struct {
	ID  int64  ___json:"id" bson:"_id"___
	Tag string ___json:"tag" bson:"tag"___
}`

	tpl = strings.ReplaceAll(tpl, "___", "`")

	return g.replaceModulePath(tpl)
}

// templateModelsUserState is the template for lib/bot/models/userstate.go file
func (g *Generator) templateModelsUserState() string {
	tpl := `package models

type UserState struct {
	ID    int64  ___bson:"_id"___
	State string ___bson:"state"___
}`

	tpl = strings.ReplaceAll(tpl, "___", "`")

	return g.replaceModulePath(tpl)
}

// templateBot is the template for lib/bot/bot.go file
func (g *Generator) templateBot() string {
	tpl := `package bot

import (
	"github.com/aliforever/go-telegram-bot-api"
	"github.com/aliforever/go-telejoon"
	"{{MODULE_PATH}}/lib/bot/db"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	api            *tgbotapi.TelegramBot
	repository     *db.Repository
	languageConfig *telejoon.LanguageConfig
	logGroupID     int64
	logger         *logrus.Logger
}

func NewBot(
	api *tgbotapi.TelegramBot,
	repository *db.Repository,
	languages *telejoon.Languages,
	logGroupID int64,
	logger *logrus.Logger,
) *Bot {
	languageConfig := telejoon.NewLanguageConfig(languages, repository).
		WithChangeLanguageMenu("ChooseLanguage", true)

	return &Bot{
		api:            api,
		repository:     repository,
		logGroupID:     logGroupID,
		languageConfig: languageConfig,
		logger:         logger,
	}
}

func (b *Bot) NewProcessor() *telejoon.EngineWithPrivateStateHandlers {
	var options []*telejoon.Options

	if b.logGroupID != 0 {
		options = append(options, telejoon.NewOptions().SetErrorGroupID(b.logGroupID))
	}

	processor := telejoon.WithPrivateStateHandlers(
		b.repository,
		"Welcome",
		options...,
	).WithLanguageConfig(b.languageConfig)

	processor.AddStaticMenu("Welcome", b.Welcome())

	return processor
}

func (b *Bot) Start() {
	for update := range b.api.Updates() {
		b.NewProcessor().Process(b.api, update)
	}
}
`

	return g.replaceModulePath(tpl)
}

// templateWelcome is the template for lib/bot/welcome.go file
func (g *Generator) templateWelcome() string {
	tpl := `package bot

import (
	"github.com/aliforever/go-telejoon"
)

func (b *Bot) Welcome() *telejoon.StaticMenu {
	actionHandlers := telejoon.NewStaticActionBuilder().
		AddStateButton(telejoon.NewLanguageKeyText("Welcome.ChangeLanguageButton"), "ChooseLanguage").
		SetButtonFormation(1).
		SetMaxButtonPerRow(2)

	return telejoon.NewStaticMenu(telejoon.NewLanguageKeyText("Welcome.Main"), actionHandlers)
}`

	return tpl
}

func (g *Generator) replaceModulePath(tpl string) string {
	return strings.ReplaceAll(tpl, "{{MODULE_PATH}}", g.ModulePath)
}

func (g *Generator) replaceBotName(tpl string) string {
	return strings.ReplaceAll(tpl, "{{BOT_NAME}}", g.botName)
}
