package application

import (
	"errors"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/commands/about"
	"github.com/kaellybot/kaelly-discord/commands/config"
	"github.com/kaellybot/kaelly-discord/commands/job"
	"github.com/kaellybot/kaelly-discord/commands/pos"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/repositories/areas"
	"github.com/kaellybot/kaelly-discord/repositories/dimensions"
	guildRepo "github.com/kaellybot/kaelly-discord/repositories/guilds"
	"github.com/kaellybot/kaelly-discord/repositories/jobs"
	serverRepo "github.com/kaellybot/kaelly-discord/repositories/servers"
	"github.com/kaellybot/kaelly-discord/repositories/subareas"
	"github.com/kaellybot/kaelly-discord/repositories/transports"
	"github.com/kaellybot/kaelly-discord/services/books"
	"github.com/kaellybot/kaelly-discord/services/discord"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/portals"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	ErrCannotInstantiateApp = errors.New("Cannot instantiate application")
)

type ApplicationInterface interface {
	Run() error
	Shutdown()
}

type Application struct {
	db             databases.MySQLConnection
	broker         amqp.MessageBrokerInterface
	guildService   guilds.GuildService
	bookService    books.BookService
	portalService  portals.PortalService
	serverService  servers.ServerService
	discordService discord.DiscordService
	commands       []commands.Command
	requestManager requests.RequestManager
}

func New() (*Application, error) {
	db, err := databases.New()
	if err != nil {
		log.Fatal().Err(err).Msgf("DB instantiation failed, shutting down.")
	}

	broker, err := amqp.New(constants.GetRabbitMQClientId(), viper.GetString(constants.RabbitMqAddress), getBindings())
	if err != nil {
		log.Fatal().Err(err).Msgf("Broker instantiation failed, shutting down.")
	}

	dimensionRepo := dimensions.New(db)
	areaRepo := areas.New(db)
	subAreaRepo := subareas.New(db)
	transportTypeRepo := transports.New(db)
	portalService, err := portals.New(dimensionRepo, areaRepo, subAreaRepo, transportTypeRepo)
	if err != nil {
		log.Fatal().Err(err).Msgf("Dimension Service instantiation failed, shutting down.")
	}

	serverRepo := serverRepo.New(db)
	serverService, err := servers.New(serverRepo)
	if err != nil {
		log.Fatal().Err(err).Msgf("Server Service instantiation failed, shutting down.")
	}

	jobRepo := jobs.New(db)
	bookService, err := books.New(jobRepo)
	if err != nil {
		log.Fatal().Err(err).Msgf("Book Service instantiation failed, shutting down.")
	}

	guildRepo := guildRepo.New(db)
	guildService := guilds.New(guildRepo)

	requestsManager := requests.New(broker)
	commands := []commands.Command{
		about.New(),
		config.New(guildService, serverService, requestsManager),
		job.New(bookService, guildService, serverService, requestsManager),
		pos.New(guildService, portalService, serverService, requestsManager),
	}

	discordService, err := discord.New(
		viper.GetString(constants.Token),
		viper.GetInt(constants.ShardId),
		viper.GetInt(constants.ShardCount),
		commands)
	if err != nil {
		return nil, err
	}

	return &Application{
		db:             db,
		broker:         broker,
		requestManager: requestsManager,
		bookService:    bookService,
		guildService:   guildService,
		portalService:  portalService,
		serverService:  serverService,
		discordService: discordService,
		commands:       commands,
	}, nil
}

func (app *Application) Run() error {
	err := app.requestManager.Listen()
	if err != nil {
		return err
	}

	err = app.discordService.Listen()
	if err != nil {
		return err
	}

	err = app.discordService.RegisterCommands()
	if err != nil {
		return err
	}

	return nil
}

func (app *Application) Shutdown() {
	app.db.Shutdown()
	app.broker.Shutdown()
	app.discordService.Shutdown()
	log.Info().Msgf("Application is no longer running")
}

func getBindings() []amqp.Binding {
	return []amqp.Binding{
		requests.GetBinding(),
	}
}
