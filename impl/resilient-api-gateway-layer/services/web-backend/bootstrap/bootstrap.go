package bootstrap

import (
	"github.com/gin-gonic/gin"

	sharedMiddleware "shared/middleware"
)

// App holds all initialized application components ready to be wired into the router.
type App struct {
	Services       *Services
	DomainServices *DomainServices
	Infrastructure *Infrastructure
	Repositories   *Repositories
	AuthMW         sharedMiddleware.AuthMiddleware
}

// Boot initializes the full application stack and returns an App ready for routing.
func Boot() *App {
	infra := initInfrastructure()
	repos := initRepositories(infra)
	svc := initDomainServices(repos)
	ucs := initUseCases(infra, repos, svc)
	ctrls := initControllers(infra, ucs)

	return &App{
		Infrastructure: infra,
		Repositories:   repos,
		DomainServices: svc,
		Services: &Services{
			AuthCtrl:     ctrls.authCtrl,
			ProductCtrl:  ctrls.productCtrl,
			CartCtrl:     ctrls.cartCtrl,
			CheckoutCtrl: ctrls.checkoutCtrl,
			ProfileCtrl:  ctrls.profileCtrl,
		},
		AuthMW: ctrls.authMW,
	}
}

// SetupEngine creates a Gin engine with the health check endpoint and API routes.
func (app *App) SetupEngine() *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "web-backend"})
	})

	SetupRoutes(r, app.Services, app.AuthMW)

	return r
}
