/**
 * api.go - rest api implementation
 *
 * @author Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
 */
package api

import (
	"../config"
	"../logging"
	"gopkg.in/gin-gonic/gin.v1"
	"gopkg.in/gin-contrib/cors.v1"
)

/* gin app */
var app *gin.Engine

/**
 * Initialize module
 */
func init() {
	gin.SetMode(gin.ReleaseMode)
}

/**
 * Starts REST API server
 */
func Start(cfg config.ApiConfig) {

	var log = logging.For("api")

	if !cfg.Enabled {
		log.Info("API disabled")
		return
	}

	log.Info("Starting up API")

	app = gin.New()

	if cfg.Cors {
		corsConfig := cors.DefaultConfig();
		corsConfig.AllowAllOrigins = true;
		corsConfig.AllowCredentials = true;
		corsConfig.AllowMethods = []string{"PUT", "POST", "DELETE", "GET", "OPTIONS"}

		app.Use(cors.New(corsConfig));
		log.Info("API CORS enabled");
	}



	r := app.Group("/")

	if cfg.BasicAuth != nil {
		log.Info("Using HTTP Basic Auth")
		r.Use(gin.BasicAuth(gin.Accounts{
			cfg.BasicAuth.Login: cfg.BasicAuth.Password,
		}))
	}

	/* attach endpoints */
	attachRoot(r)
	attachServers(r)

	var err error
	/* start rest api server */
	if cfg.Tls != nil {
		log.Info("Starting HTTPS server ", cfg.Bind)
		err = app.RunTLS(cfg.Bind, cfg.Tls.CertPath, cfg.Tls.KeyPath)
	} else {
		log.Info("Starting HTTP server ", cfg.Bind)
		err = app.Run(cfg.Bind)
	}

	if err != nil {
		log.Fatal(err)
	}

}
