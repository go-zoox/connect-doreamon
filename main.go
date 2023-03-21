package main

import (
	"log"
	"net/url"
	"strings"

	"github.com/go-zoox/core-utils/cast"
	"github.com/go-zoox/core-utils/fmt"
	"github.com/go-zoox/core-utils/regexp"
	"github.com/go-zoox/random"

	"github.com/go-zoox/cli"
	internal "github.com/go-zoox/connect/app"
	"github.com/go-zoox/connect/app/config"
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
			&cli.StringFlag{
				Name:    "frontend",
				Usage:   "frontend service",
				EnvVars: []string{"FRONTEND"},
			},
			&cli.StringFlag{
				Name:    "backend",
				Usage:   "backend service",
				EnvVars: []string{"BACKEND"},
			},
			&cli.StringFlag{
				Name:    "upstream",
				Usage:   "upstream service",
				EnvVars: []string{"UPSTREAM"},
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
		frontend := c.String("frontend")
		backend := c.String("backend")
		upstream := c.String("upstream")
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
		if sessionKey != "" {
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

		if cfg.Upstream.Host != "" && cfg.Upstream.Port != 0 {
			// ignore
		} else if upstream != "" {
			if regexp.Match("://", upstream) {
				u, err := url.Parse(upstream)
				if err != nil {
					return fmt.Errorf("upstream format error, protocol://host:port")
				}

				cfg.Upstream = config.ConfigPartService{
					Protocol: u.Scheme,
					Host:     u.Hostname(),
					Port:     cast.ToInt64(u.Port()),
				}
			} else {
				parts := strings.Split(upstream, ":")
				if len(parts) != 2 {
					return fmt.Errorf("upstream format error, host:port")
				}

				cfg.Upstream = config.ConfigPartService{
					Protocol: "http",
					Host:     parts[0],
					Port:     cast.ToInt64(parts[1]),
				}
			}
		} else {
			if frontend == "" || backend == "" {
				return fmt.Errorf("frontend and backend are required")
			}

			{
				if regexp.Match("://", frontend) {
					u, err := url.Parse(frontend)
					if err != nil {
						return fmt.Errorf("frontend format error, protocol://host:port")
					}

					cfg.Frontend = config.ConfigPartService{
						Protocol: u.Scheme,
						Host:     u.Hostname(),
						Port:     cast.ToInt64(u.Port()),
					}
				} else {
					parts := strings.Split(frontend, ":")
					if len(parts) != 2 {
						return fmt.Errorf("frontend format error, host:port")
					}

					cfg.Frontend = config.ConfigPartService{
						Host: parts[0],
						Port: cast.ToInt64(parts[1]),
					}
				}
			}

			{
				if regexp.Match("://", backend) {
					u, err := url.Parse(backend)
					if err != nil {
						return fmt.Errorf("backend format error, protocol://host:port")
					}

					cfg.Backend = config.ConfigPartService{
						Protocol: u.Scheme,
						Host:     u.Hostname(),
						Port:     cast.ToInt64(u.Port()),
					}
				} else {
					parts := strings.Split(backend, ":")
					if len(parts) != 2 {
						return fmt.Errorf("backend format error, host:port")
					}

					cfg.Backend = config.ConfigPartService{
						Host: parts[0],
						Port: cast.ToInt64(parts[1]),
					}
				}
			}
		}

		if cfg.SecretKey == "" {
			cfg.SecretKey = random.String(10)
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
