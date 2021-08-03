package application

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/AliRostami1/baagh/pkg/config"
	"github.com/AliRostami1/baagh/pkg/controller/gpio"
	"github.com/AliRostami1/baagh/pkg/database"
	"github.com/AliRostami1/baagh/pkg/logger"
	"github.com/AliRostami1/baagh/pkg/signal"
)

type Application struct {
	Log      *logger.Logger
	Config   *config.Config
	DB       *database.DB
	Gpio     *gpio.GPIO
	Ctx      context.Context
	Path     string
	Shutdown func(string)
	Cleanup  func() error
}

func New() (*Application, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := userHomeDir + "/.baagh/"

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
		ConfigPaths: []string{"/etc/baagh/", "$HOME/.baagh", "."},
	})
	// we temporarily ignore this check so it doesn't terminate the program, untill we add config support
	if err != nil {
		return nil, err
	}

	// here we are handling terminate signals
	signal.ShutdownHandler(shutdown)

	// Connect to and Initialize a db instnace
	db, err := database.New(ctx, &database.Options{
		Path:   filepath.Join(path, "badger"),
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to db: %v", err)
	}

	// initialize rpio package and allocate memory
	gpio, err := gpio.New(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("there was a problem initiating the gpio controller: %v", err)
	}

	cleanup := func() error {
		db.Close()
		gpio.Cleanup()
		return nil
	}

	return &Application{
		Log:      logger,
		Config:   config,
		DB:       db,
		Gpio:     gpio,
		Ctx:      ctx,
		Path:     path,
		Shutdown: shutdown,
		Cleanup:  cleanup,
	}, nil
}
