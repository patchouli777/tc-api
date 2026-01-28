package main

import (
	"fmt"
	"os"
	"twitchy-api/internal/app"
)

//	@title			Swagger Example API
//	@version		1.0
//	@description	desc
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8090
//	@BasePath	/api

//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	err := app.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "server shutdown: %v", err)
	}
}
