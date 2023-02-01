package main

import (
	"log"

	"github.com/go-zoox/core-utils/fmt"

	"github.com/go-zoox/cli"
	internal "github.com/go-zoox/connect/app"
	"github.com/go-zoox/connect/app/config"
	"github.com/go-zoox/random"
)

func main() {
	app := cli.NewSingleProgram(&cli.SingleProgramConfig{
		Name:        "Serve",
		Usage:       "The Serve",
		Description: "Server static files",
		// Version:     Version,
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name:    "port",
				Value:   8080,
				Usage:   "The port to listen on",
				Aliases: []string{"p"},
				EnvVars: []string{"PORT"},
			},
			&cli.StringFlag{
				Name:    "config",
				Usage:   "The config file",
				Aliases: []string{"c"},
				EnvVars: []string{"CONFIG"},
			},
			&cli.StringFlag{
				Name:    "session-key",
				Usage:   "Session Key",
				EnvVars: []string{"SESSION_KEY"},
				Value:   random.String(10),
			},
			&cli.Int64Flag{
				Name:    "session-max-age",
				Usage:   "Session Max Age",
				EnvVars: []string{"SESSION_MAX_AGE"},
				Value:   86400000,
			},
			&cli.StringFlag{
				Name:    "client-id",
				Usage:   "Doreamon Client ID",
				EnvVars: []string{"CLIENT_ID"},
			},
			&cli.StringFlag{
				Name:    "client-secret",
				Usage:   "Doreamon Client Secret",
				EnvVars: []string{"CLIENT_SECRET"},
			},
			&cli.StringFlag{
				Name:    "redirect-uri",
				Usage:   "Doreamon Client Secret",
				EnvVars: []string{"REDIRECT_URI"},
			},
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Debug mode show config info",
				EnvVars: []string{"DEBUG"},
				Value:   false,
			},
		},
	})

	app.Command(func(c *cli.Context) error {
		port := c.Int64("port")
		configFile := c.String("config")
		sessionKey := c.String("session-key")
		sessionMaxAge := c.Int64("session-max-age")
		clientID := c.String("client-id")
		clientSecret := c.String("client-secret")
		redirectURI := c.String("redirect-uri")
		debug := c.Bool("debug")

		var cfg *config.Config
		if configFile != "" {
			var err error
			if cfg, err = config.Load(configFile); err != nil {
				log.Fatal(fmt.Errorf("failed to load config (%s, %s)", configFile, err))
			}
		} else {
			cfg = &config.Config{}
		}

		//
		cfg.Auth.Mode = "oauth2"
		cfg.Auth.Provider = "doreamon"
		cfg.Services.App.Mode = "service"
		cfg.Services.App.Service = "https://api.zcorky.com/oauth/app"
		cfg.Services.User.Mode = "service"
		cfg.Services.User.Service = "https://api.zcorky.com/user"
		cfg.Services.Menus.Mode = "service"
		cfg.Services.Menus.Service = "https://api.zcorky.com/menus"
		cfg.Services.Users.Mode = "service"
		cfg.Services.Users.Service = "https://api.zcorky.com/users"
		cfg.Services.OpenID.Mode = "service"
		cfg.Services.OpenID.Service = "https://api.zcorky.com/oauth/app/user/open_id"
		if cfg.Port == 0 {
			cfg.Port = port
		}
		if cfg.SecretKey == "" {
			cfg.SecretKey = sessionKey
		}
		if cfg.SessionMaxAge == 0 {
			cfg.SessionMaxAge = sessionMaxAge
		}
		if clientID != "" {
			if clientSecret == "" || redirectURI == "" {
				return fmt.Errorf("client_id, client_secret, redirect_uri are required (1)")
			}

			cfg.OAuth2 = []config.ConfigPartAuthOAuth2{
				{
					Name:         "doreamon",
					ClientID:     clientID,
					ClientSecret: clientSecret,
					RedirectURI:  redirectURI,
				},
			}
		}

		if debug {
			fmt.PrintJSON("config:", cfg)
		}

		if len(cfg.OAuth2) == 0 {
			return fmt.Errorf("client_id, client_secret, redirect_uri are required (2)")
		}

		app := internal.New()
		if err := app.Start(cfg); err != nil {
			log.Fatal(fmt.Errorf("failed to start server(err: %s)", err))
		}

		return nil
	})

	app.Run()
}
