package mysql

// @import { ConfigurationProperties } from "github.com/spice-framework/spice/annotation/core"

// Settings defines the explicit MySQL profile configuration.
//
// @ConfigurationProperties(prefix="petclinic.datasource")
type Settings struct {
	URL           string `spice:"url,required,env=SPICE_PETCLINIC_MYSQL_URL,secret"`
	AllowInsecure bool   `spice:"allow-insecure,default=false,env=SPICE_PETCLINIC_MYSQL_ALLOW_INSECURE"`
}
