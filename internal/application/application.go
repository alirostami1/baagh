package application

import (
	"context"
	"log"

	"github.com/AliRostami1/baagh/pkg/config"
	"github.com/AliRostami1/baagh/pkg/logger"
	"github.com/AliRostami1/baagh/pkg/signal"
)

type Application struct {
	Log    *logger.Logger
	Config *config.Config
	// DB       *database.DB
	Ctx      context.Context
	Shutdown func(string)
	Cleanup  func() error
}

func New() (*Application, error) {
	// this is the application context, it will determine when the application will exit
	ctx, cancelCtx := context.WithCancel(context.Background())

	// calling shutdown will terminate the program
	shutdown := func(reason string) {
		log.Println(reason)
		cancelCtx()
	}

	// get the logger
	logger, err := logger.New(shutdown)
	if err != nil {
		return nil, err
	}

	// get the config
	config, err := config.New(&config.ConfigOptions{
		ConfigName:  "config",
		ConfigType:  "yaml",
		ConfigPaths: []string{"/etc/baagh/"},
	})
	// we temporarily ignore this check so it doesn't terminate the program, untill we add config support
	if err != nil {
		return nil, err
	}

	// here we are handling terminate signals
	signal.ShutdownHandler(shutdown)

	// // Connect to and Initialize a db instnace
	// db, err := database.New(ctx, &database.Options{
	// 	Path:   filepath.Join("/var/log/baagh/badger"),
	// 	Logger: logger,
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("couldn't connect to db: %v", err)
	// }

	cleanup := func() (err error) {
		// err := db.Close()
		// if err != nil {
		// 	logger.Errorf("problem while closing the db: %v", err)
		// }
		return
	}

	return &Application{
		Log:    logger,
		Config: config,
		// DB:       db,
		Ctx:      ctx,
		Shutdown: shutdown,
		Cleanup:  cleanup,
	}, nil
}
